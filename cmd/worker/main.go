package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"

	"github.com/smaranbhupathi/pingr/internal/adapters/outbound/checker"
	"github.com/smaranbhupathi/pingr/internal/adapters/outbound/email"
	"github.com/smaranbhupathi/pingr/internal/adapters/outbound/postgres"
	"github.com/smaranbhupathi/pingr/internal/core/ports/outbound"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file, using environment variables")
	}

	db, err := pgxpool.New(context.Background(), mustEnv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("connect db: %v", err)
	}
	defer db.Close()

	if err := db.Ping(context.Background()); err != nil {
		log.Fatalf("ping db: %v", err)
	}
	log.Println("database connected")

	region := envOr("WORKER_REGION", "us-east")

	// Wire repositories
	monitorRepo := postgres.NewMonitorRepository(db)
	checkRepo := postgres.NewCheckRepository(db)
	incidentRepo := postgres.NewIncidentRepository(db)
	alertChannelRepo := postgres.NewAlertChannelRepository(db)

	// Checkers — add TCPChecker, DNSChecker here in Roll-out 2, zero other changes
	checkers := []outbound.Checker{
		checker.NewHTTPChecker(),
	}

	// Notifiers — add SlackNotifier, DiscordNotifier here in Roll-out 2, zero other changes
	notifiers := []outbound.Notifier{
		email.NewNotifier(mustEnv("RESEND_API_KEY"), mustEnv("FROM_EMAIL")),
	}

	w := checker.NewWorker(
		region,
		monitorRepo,
		checkRepo,
		incidentRepo,
		alertChannelRepo,
		checkers,
		notifiers,
		20, // max concurrent checks — tune based on server capacity
	)

	ctx, cancel := context.WithCancel(context.Background())

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		log.Println("shutting down worker...")
		cancel()
	}()

	w.Run(ctx) // blocks until ctx cancelled
	log.Println("worker stopped")
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
