package outbound

import (
	"context"

	"github.com/smaranbhupathi/pingr/internal/core/domain"
)

// Notifier is the single interface all alert channels implement.
// Roll-out 1: EmailNotifier
// Roll-out 2: SlackNotifier, DiscordNotifier, TelegramNotifier
// Adding a new channel = implement this interface + register in main.go. Nothing else changes.
type Notifier interface {
	// Type returns the channel type this notifier handles.
	Type() domain.AlertChannelType

	// Send dispatches the alert. Config is the channel-specific config map from AlertChannel.
	Send(ctx context.Context, event domain.AlertEvent, config map[string]any) error
}

// EmailSender is used for transactional emails (verify, reset password).
// Separate from Notifier because these are system emails, not user-configured alerts.
type EmailSender interface {
	SendVerification(ctx context.Context, toEmail, token string) error
	SendPasswordReset(ctx context.Context, toEmail, token string) error
}
