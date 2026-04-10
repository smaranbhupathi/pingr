package domain

import (
	"time"

	"github.com/google/uuid"
)

type MonitorStatus string

const (
	MonitorStatusUp      MonitorStatus = "up"
	MonitorStatusDown    MonitorStatus = "down"
	MonitorStatusPaused  MonitorStatus = "paused"
	MonitorStatusPending MonitorStatus = "pending" // never checked yet
)

type MonitorType string

const (
	MonitorTypeHTTP MonitorType = "http" // Roll-out 1
	// MonitorTypeTCP  MonitorType = "tcp"  // Roll-out 2
	// MonitorTypeDNS  MonitorType = "dns"  // Roll-out 2
)

type Monitor struct {
	ID               uuid.UUID     `json:"id"`
	UserID           uuid.UUID     `json:"user_id"`
	Name             string        `json:"name"`
	URL              string        `json:"url"`
	Type             MonitorType   `json:"type"`
	IntervalSeconds  int           `json:"interval_seconds"`
	TimeoutSeconds   int           `json:"timeout_seconds"`
	FailureThreshold int           `json:"failure_threshold"`
	Region           string        `json:"region"`
	IsActive         bool          `json:"is_active"`
	Status           MonitorStatus `json:"status"`
	LastCheckedAt    *time.Time    `json:"last_checked_at"`
	CreatedAt        time.Time     `json:"created_at"`
	UpdatedAt        time.Time     `json:"updated_at"`
}

type MonitorCheck struct {
	ID             uuid.UUID `json:"id"`
	MonitorID      uuid.UUID `json:"monitor_id"`
	CheckedAt      time.Time `json:"checked_at"`
	IsUp           bool      `json:"is_up"`
	StatusCode     *int      `json:"status_code"`
	ResponseTimeMs int64     `json:"response_time_ms"`
	ErrorMessage   string    `json:"error_message"`
	Region         string    `json:"region"`
}

type Incident struct {
	ID         uuid.UUID      `json:"id"`
	MonitorID  uuid.UUID      `json:"monitor_id"`
	StartedAt  time.Time      `json:"started_at"`
	ResolvedAt *time.Time     `json:"resolved_at"`
	Duration   *time.Duration `json:"-"`
}
