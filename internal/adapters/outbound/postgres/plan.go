package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/smaranbhupathi/pingr/internal/core/domain"
	"github.com/smaranbhupathi/pingr/internal/core/ports/outbound"
)

type planRepo struct{ db *pgxpool.Pool }

func NewPlanRepository(db *pgxpool.Pool) outbound.PlanRepository {
	return &planRepo{db: db}
}

func (r *planRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Plan, error) {
	var p domain.Plan
	err := r.db.QueryRow(ctx,
		`SELECT id, name, max_monitors, min_interval_seconds, created_at FROM plans WHERE id=$1`, id,
	).Scan(&p.ID, &p.Name, &p.MaxMonitors, &p.MinIntervalSeconds, &p.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("plan by id: %w", err)
	}
	return &p, nil
}

func (r *planRepo) GetByName(ctx context.Context, name string) (*domain.Plan, error) {
	var p domain.Plan
	err := r.db.QueryRow(ctx,
		`SELECT id, name, max_monitors, min_interval_seconds, created_at FROM plans WHERE name=$1`, name,
	).Scan(&p.ID, &p.Name, &p.MaxMonitors, &p.MinIntervalSeconds, &p.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("plan by name: %w", err)
	}
	return &p, nil
}
