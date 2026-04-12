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

// ComponentStatus is the user-facing status shown on the public status page.
// Operators set this manually (via incident create/update) or the worker sets
// it automatically on down/recovery.
type ComponentStatus string

const (
	ComponentStatusOperational         ComponentStatus = "operational"
	ComponentStatusDegradedPerformance ComponentStatus = "degraded_performance"
	ComponentStatusPartialOutage       ComponentStatus = "partial_outage"
	ComponentStatusMajorOutage         ComponentStatus = "major_outage"
	ComponentStatusUnderMaintenance    ComponentStatus = "under_maintenance"
)

// Component groups monitors together on the status page (e.g. "API", "Database").
// Monitors without a component_id are shown ungrouped.
type Component struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	SortOrder   int       `json:"sort_order"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Monitor struct {
	ID               uuid.UUID       `json:"id"`
	UserID           uuid.UUID       `json:"user_id"`
	Name             string          `json:"name"`
	Description      string          `json:"description"`
	URL              string          `json:"url"`
	Type             MonitorType     `json:"type"`
	IntervalSeconds  int             `json:"interval_seconds"`
	TimeoutSeconds   int             `json:"timeout_seconds"`
	FailureThreshold int             `json:"failure_threshold"`
	Region           string          `json:"region"`
	IsActive         bool            `json:"is_active"`
	Status           MonitorStatus   `json:"status"`
	ComponentStatus  ComponentStatus `json:"component_status"`
	ComponentID      *uuid.UUID      `json:"component_id"`
	LastCheckedAt    *time.Time      `json:"last_checked_at"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
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

// OutageEvent is created automatically by the worker when a monitor goes down.
// DailyUptimeStat holds the uptime percentage for a single calendar day.
// Uptime == -1 means no checks ran that day (new monitor or gap in data).
type DailyUptimeStat struct {
	Date   string  `json:"date"`   // "2026-04-11"
	Uptime float64 `json:"uptime"` // 0–100 or -1
}

// Used purely for uptime math and alert triggering — not directly shown in the UI.
type OutageEvent struct {
	ID         uuid.UUID      `json:"id"`
	MonitorID  uuid.UUID      `json:"monitor_id"`
	StartedAt  time.Time      `json:"started_at"`
	ResolvedAt *time.Time     `json:"resolved_at"`
	Duration   *time.Duration `json:"-"`
}

// IncidentStatus represents the lifecycle stage of an incident.
type IncidentStatus string

const (
	IncidentStatusInvestigating IncidentStatus = "investigating"
	IncidentStatusIdentified    IncidentStatus = "identified"
	IncidentStatusMonitoring    IncidentStatus = "monitoring"
	IncidentStatusResolved      IncidentStatus = "resolved"
)

// Incident is a user-facing communication object shown on the public status page.
// It can be created manually by the operator or auto-seeded by the worker.
type Incident struct {
	ID            uuid.UUID        `json:"id"`
	UserID        uuid.UUID        `json:"user_id"`
	Name          string           `json:"name"`
	Status        IncidentStatus   `json:"status"`
	Source        string           `json:"source"` // "manual" | "auto"
	OutageEventID *uuid.UUID       `json:"outage_event_id,omitempty"` // set for auto-created incidents
	ResolvedAt    *time.Time       `json:"resolved_at"`
	CreatedAt     time.Time        `json:"created_at"`
	UpdatedAt     time.Time        `json:"updated_at"`
	Updates       []IncidentUpdate  `json:"updates,omitempty"`
	MonitorIDs    []uuid.UUID       `json:"monitor_ids,omitempty"`
	Monitors      []IncidentMonitor `json:"monitors,omitempty"`
}

// IncidentMonitor is a lightweight summary of a monitor attached to an incident.
type IncidentMonitor struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
	URL  string    `json:"url"`
}

// IncidentUpdate is one entry in the incident timeline.
type IncidentUpdate struct {
	ID         uuid.UUID      `json:"id"`
	IncidentID uuid.UUID      `json:"incident_id"`
	Status     IncidentStatus `json:"status"`
	Message    string         `json:"message"`
	Notify     bool           `json:"notify"`
	Source     string         `json:"source"` // "manual" | "auto"
	CreatedAt  time.Time      `json:"created_at"`
}
