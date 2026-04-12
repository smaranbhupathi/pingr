package inbound

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/smaranbhupathi/pingr/internal/core/domain"
)

type CreateMonitorInput struct {
	UserID          uuid.UUID
	Name            string
	URL             string
	IntervalSeconds int
}

type UpdateMonitorInput struct {
	Name            *string `json:"name"`
	IntervalSeconds *int    `json:"interval_seconds"`
	IsActive        *bool   `json:"is_active"`
	ComponentID     *uuid.UUID `json:"component_id"`
}

type UptimeStats struct {
	Last24h float64 `json:"last_24h"`
	Last7d  float64 `json:"last_7d"`
	Last30d float64 `json:"last_30d"`
	Last90d float64 `json:"last_90d"`
}

type MonitorDetail struct {
	Monitor        domain.Monitor          `json:"monitor"`
	Uptime         UptimeStats             `json:"uptime"`
	DailyUptime    []domain.DailyUptimeStat `json:"daily_uptime"`
	RecentCheck    *domain.MonitorCheck    `json:"recent_check"`
	Incidents      []domain.Incident       `json:"incidents"`
	ActiveIncident *domain.Incident        `json:"active_incident"`
}

type CheckDataPoint struct {
	Timestamp      time.Time `json:"timestamp"`
	ResponseTimeMs int64     `json:"response_time_ms"`
	IsUp           bool      `json:"is_up"`
}

type MonitorService interface {
	Create(ctx context.Context, input CreateMonitorInput) (*domain.Monitor, error)
	GetByID(ctx context.Context, id, userID uuid.UUID) (*MonitorDetail, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]domain.Monitor, error)
	Update(ctx context.Context, id, userID uuid.UUID, input UpdateMonitorInput) (*domain.Monitor, error)
	Delete(ctx context.Context, id, userID uuid.UUID) error
	GetResponseTimeGraph(ctx context.Context, monitorID uuid.UUID, from, to time.Time) ([]CheckDataPoint, error)

	// Public — no auth required
	GetStatusPage(ctx context.Context, username string) ([]MonitorDetail, error)
}
