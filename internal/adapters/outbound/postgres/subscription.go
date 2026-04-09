package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/smaranbhupathi/pingr/internal/core/domain"
	"github.com/smaranbhupathi/pingr/internal/core/ports/outbound"
	"github.com/google/uuid"
)

type alertSubRepo struct{ db *pgxpool.Pool }

func NewAlertSubscriptionRepository(db *pgxpool.Pool) outbound.AlertSubscriptionRepository {
	return &alertSubRepo{db: db}
}

func (r *alertSubRepo) Create(ctx context.Context, sub *domain.AlertSubscription) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO alert_subscriptions (id, monitor_id, alert_channel_id, created_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (monitor_id, alert_channel_id) DO NOTHING`,
		sub.ID, sub.MonitorID, sub.AlertChannelID, sub.CreatedAt,
	)
	return err
}

func (r *alertSubRepo) DeleteByMonitorAndChannel(ctx context.Context, monitorID, channelID uuid.UUID) error {
	_, err := r.db.Exec(ctx,
		`DELETE FROM alert_subscriptions WHERE monitor_id=$1 AND alert_channel_id=$2`,
		monitorID, channelID,
	)
	return err
}
