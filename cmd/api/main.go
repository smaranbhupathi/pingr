package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	inboundhttp "github.com/smaranbhupathi/pingr/internal/adapters/inbound/http"
	"github.com/smaranbhupathi/pingr/internal/adapters/inbound/http/handler"
	"github.com/smaranbhupathi/pingr/internal/adapters/outbound/email"
	"github.com/smaranbhupathi/pingr/internal/adapters/outbound/postgres"
	"github.com/smaranbhupathi/pingr/internal/core/services"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file, using environment variables")
	}

	db, err := postgres.Connect(context.Background(), mustEnv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("connect db: %v", err)
	}
	defer db.Close()
	log.Println("database connected")

	// Outbound adapters
	userRepo         := postgres.NewUserRepository(db)
	planRepo         := postgres.NewPlanRepository(db)
	monitorRepo      := postgres.NewMonitorRepository(db)
	checkRepo        := postgres.NewCheckRepository(db)
	incidentRepo     := postgres.NewIncidentRepository(db)
	alertChannelRepo := postgres.NewAlertChannelRepository(db)
	alertSubRepo     := postgres.NewAlertSubscriptionRepository(db)

	emailSender := email.NewEmailSender(
		mustEnv("RESEND_API_KEY"),
		mustEnv("FROM_EMAIL"),
		mustEnv("APP_BASE_URL"),
	)

	// Core services
	authSvc := services.NewAuthService(userRepo, planRepo, alertChannelRepo, emailSender, services.AuthServiceConfig{
		JWTSecret:            mustEnv("JWT_SECRET"),
		AccessTokenDuration:  15 * time.Minute,
		RefreshTokenDuration: 7 * 24 * time.Hour,
		AppBaseURL:           mustEnv("APP_BASE_URL"),
	})
	monitorSvc := services.NewMonitorService(monitorRepo, checkRepo, incidentRepo, userRepo, planRepo)
	userSvc    := services.NewUserService(userRepo, planRepo, alertChannelRepo, alertSubRepo, monitorRepo)

	// HTTP handlers
	authH    := handler.NewAuthHandler(authSvc)
	monitorH := handler.NewMonitorHandler(monitorSvc)
	userH    := handler.NewUserHandler(userSvc)

	router := inboundhttp.NewRouter(authH, monitorH, userH, mustEnv("JWT_SECRET"))

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
		log.Printf("API server listening on :%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	<-quit
	log.Println("shutting down API server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("forced shutdown: %v", err)
	}
	log.Println("API server stopped")
}

func mustEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("required env var %s is not set", key)
	}
	return val
}

func envOr(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
