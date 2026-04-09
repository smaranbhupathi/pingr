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
	Name            *string
	IntervalSeconds *int
	IsActive        *bool
}

type UptimeStats struct {
	Last24h float64
	Last7d  float64
	Last30d float64
	Last90d float64
}

type MonitorDetail struct {
	Monitor     domain.Monitor
	Uptime      UptimeStats
	RecentCheck *domain.MonitorCheck
	Incidents   []domain.Incident
}

type CheckDataPoint struct {
	Timestamp      time.Time
	ResponseTimeMs int64
	IsUp           bool
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
