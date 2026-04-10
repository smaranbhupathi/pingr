package email

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/smaranbhupathi/pingr/internal/core/domain"
	"github.com/smaranbhupathi/pingr/internal/core/ports/outbound"
)

// NewConsoleSender returns an EmailSender that prints emails to stdout via slog.
// Use this in local development — swap for NewEmailSender in production.
func NewConsoleSender(appBaseURL string) outbound.EmailSender {
	return &consoleSender{appBaseURL: appBaseURL}
}

// NewConsoleNotifier returns a Notifier that prints alerts to stdout via slog.
func NewConsoleNotifier() outbound.Notifier {
	return &consoleNotifier{}
}

type consoleSender struct{ appBaseURL string }

func (c *consoleSender) SendVerification(ctx context.Context, toEmail, token string) error {
	link := fmt.Sprintf("%s/verify-email?token=%s", c.appBaseURL, token)
	slog.Info("📧 EMAIL — verify account",
		"to", toEmail,
		"link", link,
	)
	return nil
}

func (c *consoleSender) SendPasswordReset(ctx context.Context, toEmail, token string) error {
	link := fmt.Sprintf("%s/reset-password?token=%s", c.appBaseURL, token)
	slog.Info("📧 EMAIL — password reset",
		"to", toEmail,
		"link", link,
	)
	return nil
}

func (c *consoleSender) SendSubscriptionConfirmation(ctx context.Context, toEmail, monitorName, monitorURL string) error {
	slog.Info("📧 EMAIL — alert subscription confirmed",
		"to", toEmail,
		"monitor", monitorName,
		"url", monitorURL,
	)
	return nil
}

type consoleNotifier struct{}

func (n *consoleNotifier) Type() domain.AlertChannelType {
	return domain.AlertChannelEmail
}

func (n *consoleNotifier) Send(ctx context.Context, event domain.AlertEvent, config map[string]any) error {
	to, _ := config["email"].(string)
	slog.Warn("🚨 ALERT fired",
		"type", event.Type,
		"monitor", event.Monitor.Name,
		"url", event.Monitor.URL,
		"to", to,
	)
	return nil
}
