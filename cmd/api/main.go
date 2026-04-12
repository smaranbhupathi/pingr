package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	inboundhttp "github.com/smaranbhupathi/pingr/internal/adapters/inbound/http"
	"github.com/smaranbhupathi/pingr/internal/adapters/inbound/http/handler"
	"github.com/smaranbhupathi/pingr/internal/adapters/inbound/http/ratelimit"
	"github.com/smaranbhupathi/pingr/internal/adapters/outbound/email"
	"github.com/smaranbhupathi/pingr/internal/adapters/outbound/postgres"
	"github.com/smaranbhupathi/pingr/internal/adapters/outbound/storage"
	"github.com/smaranbhupathi/pingr/internal/adapters/outbound/webhook"
	"github.com/smaranbhupathi/pingr/internal/config"
	"github.com/smaranbhupathi/pingr/internal/core/ports/outbound"
	"github.com/smaranbhupathi/pingr/internal/core/services"
	"github.com/smaranbhupathi/pingr/internal/logger"
)

func main() {
	if err := godotenv.Load(); err != nil {
		slog.Warn("no .env file, falling back to environment variables")
	}

	env := envOr("APP_ENV", "dev")
	log := logger.New(env)

	cfg, err := config.Load()
	if err != nil {
		// Only fatal if the file exists but is malformed.
		// Missing file is handled inside Load() with a warning + defaults.
		slog.Error("failed to load config.yaml", "error", err)
		os.Exit(1)
	}
	log.Info("config loaded",
		"email_alerts", cfg.Features.EmailAlerts,
		"slack_alerts", cfg.Features.SlackAlerts,
		"discord_alerts", cfg.Features.DiscordAlerts,
	)

	log.Info("starting API server", "env", env)

	db, err := postgres.Connect(context.Background(), mustEnv("DATABASE_URL"))
	if err != nil {
		log.Error("connect db failed", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	log.Info("database connected")

	// Outbound adapters
	userRepo         := postgres.NewUserRepository(db)
	planRepo         := postgres.NewPlanRepository(db)
	monitorRepo      := postgres.NewMonitorRepository(db)
	checkRepo        := postgres.NewCheckRepository(db)
	incidentRepo     := postgres.NewIncidentRepository(db)
	alertChannelRepo := postgres.NewAlertChannelRepository(db)
	alertSubRepo     := postgres.NewAlertSubscriptionRepository(db)
	componentRepo    := postgres.NewComponentRepository(db)

	var emailSender outbound.EmailSender
	if resendKey := os.Getenv("RESEND_API_KEY"); resendKey != "" {
		emailSender = email.NewEmailSender(resendKey, mustEnv("FROM_EMAIL"), mustEnv("APP_BASE_URL"))
		log.Info("using Resend email sender", "from", mustEnv("FROM_EMAIL"))
	} else {
		emailSender = email.NewConsoleSender(mustEnv("APP_BASE_URL"))
		log.Info("RESEND_API_KEY not set — using console email sender (links printed to log)")
	}

	// Storage adapter — optional. When STORAGE_ENDPOINT is not set,
	// avatar uploads return 501 Not Implemented instead of panicking.
	// To use MinIO locally: set STORAGE_ENDPOINT=http://localhost:9000
	// To use Cloudflare R2: set STORAGE_ENDPOINT=https://<account>.r2.cloudflarestorage.com
	var storageSvc outbound.StorageService
	if storageEndpoint := os.Getenv("STORAGE_ENDPOINT"); storageEndpoint != "" {
		s, err := storage.NewS3Store(storage.Config{
			Endpoint:        storageEndpoint,
			Region:          envOr("STORAGE_REGION", "auto"),
			AccessKeyID:     mustEnv("STORAGE_ACCESS_KEY_ID"),
			SecretAccessKey: mustEnv("STORAGE_SECRET_ACCESS_KEY"),
			Bucket:          mustEnv("STORAGE_BUCKET"),
			PublicBaseURL:   mustEnv("STORAGE_PUBLIC_BASE_URL"),
		})
		if err != nil {
			log.Error("init storage failed", "error", err)
			os.Exit(1)
		}
		storageSvc = s
		log.Info("storage initialised", "endpoint", storageEndpoint, "bucket", mustEnv("STORAGE_BUCKET"))
	} else {
		log.Info("STORAGE_ENDPOINT not set — avatar uploads disabled")
	}

	// Core services
	authSvc := services.NewAuthService(userRepo, planRepo, alertChannelRepo, emailSender, services.AuthServiceConfig{
		JWTSecret:            mustEnv("JWT_SECRET"),
		AccessTokenDuration:  15 * time.Minute,
		RefreshTokenDuration: 7 * 24 * time.Hour,
		AppBaseURL:           mustEnv("APP_BASE_URL"),
	})
	monitorSvc := services.NewMonitorService(monitorRepo, checkRepo, incidentRepo, userRepo, planRepo)
	var emailNotifier outbound.Notifier
	if resendKey := os.Getenv("RESEND_API_KEY"); resendKey != "" {
		emailNotifier = email.NewNotifier(resendKey, mustEnv("FROM_EMAIL"))
	} else {
		emailNotifier = email.NewConsoleNotifier()
	}
	allNotifiers := []outbound.Notifier{
		emailNotifier,
		webhook.NewSlackNotifier(),
		webhook.NewDiscordNotifier(),
	}
	userSvc := services.NewUserService(userRepo, planRepo, alertChannelRepo, alertSubRepo, monitorRepo, componentRepo, incidentRepo, emailSender, storageSvc, allNotifiers)

	// HTTP handlers
	authH      := handler.NewAuthHandler(authSvc, log)
	monitorH   := handler.NewMonitorHandler(monitorSvc, log)
	userH      := handler.NewUserHandler(userSvc, cfg, log)
	incidentH  := handler.NewIncidentHandler(userSvc, log)
	componentH := handler.NewComponentHandler(userSvc, log)

	// Rate limiter — in-memory sliding window.
	// To switch to Redis: replace NewMemoryStore() with NewRedisStore(redisClient).
	// Everything else (router, middleware, config) stays the same.
	rlStore := ratelimit.NewMemoryStore()
	defer rlStore.Close()
	log.Info("rate limiter initialised", "store", "memory")

	allowedOrigin := envOr("ALLOWED_ORIGIN", "*")
	router := inboundhttp.NewRouter(authH, monitorH, userH, incidentH, componentH, mustEnv("JWT_SECRET"), allowedOrigin, rlStore, log)

	port := envOr("PORT", "8080")
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", port),
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Info("API server listening", "port", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	<-quit
	log.Info("shutting down API server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("forced shutdown", "error", err)
		os.Exit(1)
	}
	log.Info("API server stopped")
}

func mustEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		slog.Error("required env var not set", "key", key)
		os.Exit(1)
	}
	return val
}

func envOr(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
