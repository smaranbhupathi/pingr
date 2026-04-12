package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/smaranbhupathi/pingr/internal/core/domain"
	"github.com/smaranbhupathi/pingr/internal/core/ports/outbound"
)

type componentRepo struct{ db *pgxpool.Pool }

func NewComponentRepository(db *pgxpool.Pool) outbound.ComponentRepository {
	return &componentRepo{db: db}
}

func (r *componentRepo) Create(ctx context.Context, c *domain.Component) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO components (id, user_id, name, description, sort_order, created_at, updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		c.ID, c.UserID, c.Name, c.Description, c.SortOrder, c.CreatedAt, c.UpdatedAt,
	)
	return err
}

func (r *componentRepo) GetByID(ctx context.Context, id, userID uuid.UUID) (*domain.Component, error) {
	var c domain.Component
	err := r.db.QueryRow(ctx,
		`SELECT id, user_id, name, description, sort_order, created_at, updated_at
		 FROM components WHERE id=$1 AND user_id=$2`,
		id, userID,
	).Scan(&c.ID, &c.UserID, &c.Name, &c.Description, &c.SortOrder, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("component scan: %w", err)
	}
	return &c, nil
}

func (r *componentRepo) ListByUser(ctx context.Context, userID uuid.UUID) ([]domain.Component, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, user_id, name, description, sort_order, created_at, updated_at
		 FROM components WHERE user_id=$1 ORDER BY sort_order ASC, created_at ASC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []domain.Component
	for rows.Next() {
		var c domain.Component
		if err := rows.Scan(&c.ID, &c.UserID, &c.Name, &c.Description, &c.SortOrder, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, fmt.Errorf("component scan: %w", err)
		}
		result = append(result, c)
	}
	return result, rows.Err()
}

func (r *componentRepo) Update(ctx context.Context, c *domain.Component) error {
	_, err := r.db.Exec(ctx,
		`UPDATE components SET name=$2, description=$3, sort_order=$4, updated_at=$5 WHERE id=$1 AND user_id=$6`,
		c.ID, c.Name, c.Description, c.SortOrder, time.Now(), c.UserID,
	)
	return err
}

func (r *componentRepo) Delete(ctx context.Context, id, userID uuid.UUID) error {
	_, err := r.db.Exec(ctx,
		`DELETE FROM components WHERE id=$1 AND user_id=$2`,
		id, userID,
	)
	return err
}
