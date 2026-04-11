package domain

import (
	"time"

	"github.com/google/uuid"
)

type AlertChannelType string

const (
	AlertChannelEmail   AlertChannelType = "email"
	AlertChannelSlack   AlertChannelType = "slack"
	AlertChannelDiscord AlertChannelType = "discord"
)

// AlertChannel stores the type + all config as a JSON blob.
// Adding Slack later = new row with type="slack" and slack config in Config.
// Zero schema changes needed.
type AlertChannel struct {
	ID        uuid.UUID        `json:"id"`
	UserID    uuid.UUID        `json:"user_id"`
	Name      string           `json:"name"`
	Type      AlertChannelType `json:"type"`
	Config    map[string]any   `json:"config"`
	IsDefault bool             `json:"is_default"`
	IsEnabled bool             `json:"is_enabled"`
	CreatedAt time.Time        `json:"created_at"`
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
	Monitor     Monitor
	OutageEvent OutageEvent
	Type        AlertEventType
}
