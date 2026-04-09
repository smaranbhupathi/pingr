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
	ID               uuid.UUID
	UserID           uuid.UUID
	Name             string
	URL              string
	Type             MonitorType
	IntervalSeconds  int           // configurable per monitor, stored in DB
	TimeoutSeconds   int           // how long before a check times out
	FailureThreshold int           // consecutive failures before marking DOWN
	Region           string        // "us-east", "eu-west" etc — region-tagged from day 1
	IsActive         bool
	Status           MonitorStatus
	LastCheckedAt    *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type MonitorCheck struct {
	ID             uuid.UUID
	MonitorID      uuid.UUID
	CheckedAt      time.Time
	IsUp           bool
	StatusCode     *int
	ResponseTimeMs int64
	ErrorMessage   string
	Region         string
}

type Incident struct {
	ID         uuid.UUID
	MonitorID  uuid.UUID
	StartedAt  time.Time
	ResolvedAt *time.Time
	Duration   *time.Duration // nil if ongoing
}
