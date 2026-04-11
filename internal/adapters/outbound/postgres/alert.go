package postgres

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/smaranbhupathi/pingr/internal/core/domain"
	"github.com/smaranbhupathi/pingr/internal/core/ports/outbound"
)

type alertChannelRepo struct{ db *pgxpool.Pool }

func NewAlertChannelRepository(db *pgxpool.Pool) outbound.AlertChannelRepository {
	return &alertChannelRepo{db: db}
}

func (r *alertChannelRepo) Create(ctx context.Context, ch *domain.AlertChannel) error {
	configJSON, err := json.Marshal(ch.Config)
	if err != nil {
		return err
	}
	_, err = r.db.Exec(ctx,
		`INSERT INTO alert_channels (id, user_id, name, type, config, is_default, created_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		ch.ID, ch.UserID, ch.Name, ch.Type, configJSON, ch.IsDefault, ch.CreatedAt,
	)
	return err
}

func (r *alertChannelRepo) GetByUserID(ctx context.Context, userID uuid.UUID) ([]domain.AlertChannel, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, user_id, name, type, config, is_default, created_at FROM alert_channels WHERE user_id=$1 AND deleted_at IS NULL`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanAlertChannels(rows)
}

func (r *alertChannelRepo) GetByMonitorID(ctx context.Context, monitorID uuid.UUID) ([]domain.AlertChannel, error) {
	rows, err := r.db.Query(ctx, `
		SELECT ac.id, ac.user_id, ac.name, ac.type, ac.config, ac.is_default, ac.created_at
		FROM alert_channels ac
		JOIN alert_subscriptions s ON s.alert_channel_id = ac.id
		WHERE s.monitor_id = $1 AND ac.deleted_at IS NULL`, monitorID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanAlertChannels(rows)
}

func (r *alertChannelRepo) GetByID(ctx context.Context, id, userID uuid.UUID) (*domain.AlertChannel, error) {
	row := r.db.QueryRow(ctx,
		`SELECT id, user_id, name, type, config, is_default, created_at
		 FROM alert_channels WHERE id=$1 AND user_id=$2 AND deleted_at IS NULL`,
		id, userID,
	)
	var ch domain.AlertChannel
	var configJSON []byte
	if err := row.Scan(&ch.ID, &ch.UserID, &ch.Name, &ch.Type, &configJSON, &ch.IsDefault, &ch.CreatedAt); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(configJSON, &ch.Config); err != nil {
		return nil, err
	}
	return &ch, nil
}

func (r *alertChannelRepo) UpdateName(ctx context.Context, id, userID uuid.UUID, name string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE alert_channels SET name=$1 WHERE id=$2 AND user_id=$3 AND deleted_at IS NULL`,
		name, id, userID,
	)
	return err
}

func (r *alertChannelRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `UPDATE alert_channels SET deleted_at = NOW() WHERE id=$1 AND deleted_at IS NULL`, id)
	return err
}

func scanAlertChannels(rows interface {
	Next() bool
	Scan(...any) error
	Err() error
}) ([]domain.AlertChannel, error) {
	channels := make([]domain.AlertChannel, 0)
	for rows.Next() {
		var ch domain.AlertChannel
		var configJSON []byte
		if err := rows.Scan(&ch.ID, &ch.UserID, &ch.Name, &ch.Type, &configJSON, &ch.IsDefault, &ch.CreatedAt); err != nil {
			return nil, fmt.Errorf("alert channel scan: %w", err)
		}
		if err := json.Unmarshal(configJSON, &ch.Config); err != nil {
			return nil, err
		}
		channels = append(channels, ch)
	}
	return channels, rows.Err()
}
