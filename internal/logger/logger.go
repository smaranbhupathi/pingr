package logger

import (
	"log/slog"
	"os"

	"github.com/lmittmann/tint"
)

// New creates a structured logger.
// Dev:  colorized human-readable output via tint
// Prod: JSON — ready for log aggregators (Datadog, Loki, CloudWatch)
func New(env string) *slog.Logger {
	var handler slog.Handler
	if env == "production" {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level:     slog.LevelDebug,
			AddSource: true,
		})
	} else {
		handler = tint.NewHandler(os.Stdout, &tint.Options{
			Level:     slog.LevelDebug,
			NoColor:   false,
			AddSource: true,
		})
	}

	l := slog.New(handler)
	slog.SetDefault(l)
	return l
}
