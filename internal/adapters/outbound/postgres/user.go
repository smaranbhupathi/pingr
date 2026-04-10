package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/smaranbhupathi/pingr/internal/core/domain"
	"github.com/smaranbhupathi/pingr/internal/core/ports/outbound"
)

type userRepo struct{ db *pgxpool.Pool }

func NewUserRepository(db *pgxpool.Pool) outbound.UserRepository {
	return &userRepo{db: db}
}

func (r *userRepo) Create(ctx context.Context, u *domain.User) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO users (id, email, username, password_hash, is_verified, verify_token, plan_id, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		u.ID, u.Email, u.Username, u.PasswordHash, u.IsVerified, u.VerifyToken, u.PlanID, u.CreatedAt, u.UpdatedAt,
	)
	return err
}

func (r *userRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return r.scan(r.db.QueryRow(ctx, `SELECT `+userColumns+` FROM users WHERE id=$1`, id))
}

func (r *userRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	return r.scan(r.db.QueryRow(ctx, `SELECT `+userColumns+` FROM users WHERE email=$1`, email))
}

func (r *userRepo) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	return r.scan(r.db.QueryRow(ctx, `SELECT `+userColumns+` FROM users WHERE username=$1`, username))
}

func (r *userRepo) GetByVerifyToken(ctx context.Context, token string) (*domain.User, error) {
	return r.scan(r.db.QueryRow(ctx, `SELECT `+userColumns+` FROM users WHERE verify_token=$1`, token))
}

func (r *userRepo) GetByResetToken(ctx context.Context, token string) (*domain.User, error) {
	return r.scan(r.db.QueryRow(ctx, `SELECT `+userColumns+` FROM users WHERE reset_token=$1`, token))
}

func (r *userRepo) Update(ctx context.Context, u *domain.User) error {
	_, err := r.db.Exec(ctx, `
		UPDATE users SET
			email=$2, username=$3, password_hash=$4, is_verified=$5,
			verify_token=$6, reset_token=$7, reset_expires_at=$8,
			plan_id=$9, avatar_url=$10, updated_at=$11
		WHERE id=$1`,
		u.ID, u.Email, u.Username, u.PasswordHash, u.IsVerified,
		u.VerifyToken, u.ResetToken, u.ResetExpiresAt,
		u.PlanID, u.AvatarURL, u.UpdatedAt,
	)
	return err
}

const userColumns = `id, email, username, password_hash, is_verified, verify_token,
	COALESCE(reset_token,'') as reset_token, reset_expires_at, plan_id, avatar_url, created_at, updated_at`

func (r *userRepo) scan(row pgx.Row) (*domain.User, error) {
	var u domain.User
	err := row.Scan(
		&u.ID, &u.Email, &u.Username, &u.PasswordHash, &u.IsVerified, &u.VerifyToken,
		&u.ResetToken, &u.ResetExpiresAt, &u.PlanID, &u.AvatarURL, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("user scan: %w", err)
	}
	return &u, nil
}
