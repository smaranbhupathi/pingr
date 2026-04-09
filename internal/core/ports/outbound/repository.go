package outbound

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/smaranbhupathi/pingr/internal/core/domain"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetByUsername(ctx context.Context, username string) (*domain.User, error)
	GetByVerifyToken(ctx context.Context, token string) (*domain.User, error)
	GetByResetToken(ctx context.Context, token string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
}

type PlanRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Plan, error)
	GetByName(ctx context.Context, name string) (*domain.Plan, error)
}

type MonitorRepository interface {
	Create(ctx context.Context, monitor *domain.Monitor) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Monitor, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Monitor, error)
	GetByUsername(ctx context.Context, username string) ([]domain.Monitor, error) // for public status page
	GetDue(ctx context.Context, region string) ([]domain.Monitor, error)          // monitors due for checking
	Update(ctx context.Context, monitor *domain.Monitor) error
	Delete(ctx context.Context, id uuid.UUID) error
	CountByUserID(ctx context.Context, userID uuid.UUID) (int, error)
}

type CheckRepository interface {
	Create(ctx context.Context, check *domain.MonitorCheck) error
	GetByMonitorID(ctx context.Context, monitorID uuid.UUID, from, to time.Time) ([]domain.MonitorCheck, error)
	GetUptimeStats(ctx context.Context, monitorID uuid.UUID, from time.Time) (float64, error) // returns uptime %
	GetLatest(ctx context.Context, monitorID uuid.UUID) (*domain.MonitorCheck, error)
}

type IncidentRepository interface {
	Create(ctx context.Context, incident *domain.Incident) error
	GetOpenByMonitorID(ctx context.Context, monitorID uuid.UUID) (*domain.Incident, error)
	GetByMonitorID(ctx context.Context, monitorID uuid.UUID) ([]domain.Incident, error)
	Resolve(ctx context.Context, incidentID uuid.UUID, resolvedAt time.Time) error
}

type AlertChannelRepository interface {
	Create(ctx context.Context, channel *domain.AlertChannel) error
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]domain.AlertChannel, error)
	GetByMonitorID(ctx context.Context, monitorID uuid.UUID) ([]domain.AlertChannel, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type AlertSubscriptionRepository interface {
	Create(ctx context.Context, sub *domain.AlertSubscription) error
	DeleteByMonitorAndChannel(ctx context.Context, monitorID, channelID uuid.UUID) error
}
