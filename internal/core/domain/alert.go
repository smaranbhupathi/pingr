package domain

import (
	"time"

	"github.com/google/uuid"
)

type AlertChannelType string

const (
	AlertChannelEmail    AlertChannelType = "email"
	// AlertChannelSlack    AlertChannelType = "slack"    // Roll-out 2
	// AlertChannelDiscord  AlertChannelType = "discord"  // Roll-out 2
	// AlertChannelTelegram AlertChannelType = "telegram" // Roll-out 2
)

// AlertChannel stores the type + all config as a JSON blob.
// Adding Slack later = new row with type="slack" and slack config in Config.
// Zero schema changes needed.
type AlertChannel struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Type      AlertChannelType
	Config    map[string]any // e.g. {"email": "user@example.com"} or {"webhook_url": "..."}
	IsDefault bool
	CreatedAt time.Time
}

// AlertSubscription links a monitor to an alert channel.
type AlertSubscription struct {
	ID             uuid.UUID
	MonitorID      uuid.UUID
	AlertChannelID uuid.UUID
	CreatedAt      time.Time
}

type AlertEventType string

const (
	AlertEventDown     AlertEventType = "down"
	AlertEventRecovery AlertEventType = "recovery"
)

type AlertEvent struct {
	Monitor  Monitor
	Incident Incident
	Type     AlertEventType
}
