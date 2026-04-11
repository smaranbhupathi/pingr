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

	// SendSubscriptionConfirmation sends a one-time message when a monitor is subscribed
	// to this channel, so the user can verify the webhook is wired up correctly.
	// Implementations that don't support this (e.g. email — handled separately) return nil.
	SendSubscriptionConfirmation(ctx context.Context, monitorName, monitorURL string, config map[string]any) error

	// SendIncidentUpdate fans out a notification when the operator posts an update
	// to an incident with notify=true. Includes incident name, new status, the update
	// message, affected monitors, and timestamp.
	SendIncidentUpdate(ctx context.Context, incident domain.Incident, update domain.IncidentUpdate, config map[string]any) error
}

// EmailSender is used for transactional emails (verify, reset password, subscription confirmation).
// Separate from Notifier because these are system emails, not user-configured alerts.
type EmailSender interface {
	SendVerification(ctx context.Context, toEmail, token string) error
	SendPasswordReset(ctx context.Context, toEmail, token string) error
	SendSubscriptionConfirmation(ctx context.Context, toEmail, monitorName, monitorURL string) error
}
