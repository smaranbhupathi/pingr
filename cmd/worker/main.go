package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"

	"github.com/smaranbhupathi/pingr/internal/adapters/outbound/checker"
	"github.com/smaranbhupathi/pingr/internal/adapters/outbound/email"
	"github.com/smaranbhupathi/pingr/internal/adapters/outbound/postgres"
	"github.com/smaranbhupathi/pingr/internal/adapters/outbound/webhook"
	"github.com/smaranbhupathi/pingr/internal/config"
	"github.com/smaranbhupathi/pingr/internal/core/ports/outbound"
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
		log.Error("failed to load config.yaml", "error", err)
		os.Exit(1)
	}
	log.Info("config loaded",
		"email_alerts", cfg.Features.EmailAlerts,
		"slack_alerts", cfg.Features.SlackAlerts,
		"discord_alerts", cfg.Features.DiscordAlerts,
		"worker_tick_seconds", cfg.Monitoring.WorkerTickSeconds,
	)

	log.Info("starting worker", "env", env)

	db, err := postgres.Connect(context.Background(), mustEnv("DATABASE_URL"))
	if err != nil {
		log.Error("connect db failed", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	log.Info("database connected")

	region := envOr("WORKER_REGION", "us-east")

	// Repositories
	monitorRepo      := postgres.NewMonitorRepository(db)
	checkRepo        := postgres.NewCheckRepository(db)
	incidentRepo     := postgres.NewIncidentRepository(db)
	alertChannelRepo := postgres.NewAlertChannelRepository(db)

	// Checkers — add TCPChecker, DNSChecker here in Roll-out 2
	checkers := []outbound.Checker{
		checker.NewHTTPChecker(),
	}

	// Notifiers — email + webhook (Slack + Discord) always registered.
	// Email uses Resend if key present, otherwise logs to console.
	var emailNotifier outbound.Notifier
	if resendKey := os.Getenv("RESEND_API_KEY"); resendKey != "" {
		emailNotifier = email.NewNotifier(resendKey, mustEnv("FROM_EMAIL"))
		log.Info("using Resend notifier", "from", mustEnv("FROM_EMAIL"))
	} else {
		emailNotifier = email.NewConsoleNotifier()
		log.Info("RESEND_API_KEY not set — using console notifier (alerts printed to log)")
	}

	notifiers := []outbound.Notifier{
		emailNotifier,
		webhook.NewSlackNotifier(),
		webhook.NewDiscordNotifier(),
	}

	w := checker.NewWorker(
		region,
		monitorRepo,
		checkRepo,
		incidentRepo,
		alertChannelRepo,
		checkers,
		notifiers,
		cfg,
		20,
	)

	ctx, cancel := context.WithCancel(context.Background())

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		log.Info("shutting down worker...")
		cancel()
	}()

	w.Run(ctx)
	log.Info("worker stopped")
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
