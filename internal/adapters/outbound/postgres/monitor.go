package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/smaranbhupathi/pingr/internal/core/domain"
	"github.com/smaranbhupathi/pingr/internal/core/ports/outbound"
)

type monitorRepo struct{ db *pgxpool.Pool }

func NewMonitorRepository(db *pgxpool.Pool) outbound.MonitorRepository {
	return &monitorRepo{db: db}
}

func (r *monitorRepo) Create(ctx context.Context, m *domain.Monitor) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO monitors
			(id, user_id, name, url, type, interval_seconds, timeout_seconds,
			 failure_threshold, region, is_active, status, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
		m.ID, m.UserID, m.Name, m.URL, m.Type, m.IntervalSeconds, m.TimeoutSeconds,
		m.FailureThreshold, m.Region, m.IsActive, m.Status, m.CreatedAt, m.UpdatedAt,
	)
	return err
}

func (r *monitorRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Monitor, error) {
	return r.scanOne(r.db.QueryRow(ctx, `SELECT `+monitorCols+` FROM monitors WHERE id=$1`, id))
}

func (r *monitorRepo) GetByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Monitor, error) {
	rows, err := r.db.Query(ctx, `SELECT `+monitorCols+` FROM monitors WHERE user_id=$1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	return r.scanMany(rows)
}

func (r *monitorRepo) GetByUsername(ctx context.Context, username string) ([]domain.Monitor, error) {
	rows, err := r.db.Query(ctx, `
		SELECT `+monitorCols+` FROM monitors m
		JOIN users u ON u.id = m.user_id
		WHERE u.username=$1 AND m.is_active=true
		ORDER BY m.created_at DESC`, username,
	)
	if err != nil {
		return nil, err
	}
	return r.scanMany(rows)
}

// GetDue returns monitors that are due for checking based on their interval.
// Uses interval arithmetic so the scheduler just calls this in a tight loop.
func (r *monitorRepo) GetDue(ctx context.Context, region string) ([]domain.Monitor, error) {
	rows, err := r.db.Query(ctx, `
		SELECT `+monitorCols+` FROM monitors
		WHERE is_active = true
		  AND region = $1
		  AND (
		    last_checked_at IS NULL
		    OR last_checked_at <= NOW() - (interval_seconds || ' seconds')::interval
		  )
		ORDER BY last_checked_at ASC NULLS FIRST
		LIMIT 100`, region,
	)
	if err != nil {
		return nil, err
	}
	return r.scanMany(rows)
}

func (r *monitorRepo) Update(ctx context.Context, m *domain.Monitor) error {
	_, err := r.db.Exec(ctx, `
		UPDATE monitors SET
			name=$2, url=$3, interval_seconds=$4, timeout_seconds=$5,
			failure_threshold=$6, is_active=$7, status=$8,
			last_checked_at=$9, updated_at=$10
		WHERE id=$1`,
		m.ID, m.Name, m.URL, m.IntervalSeconds, m.TimeoutSeconds,
		m.FailureThreshold, m.IsActive, m.Status, m.LastCheckedAt, m.UpdatedAt,
	)
	return err
}

func (r *monitorRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM monitors WHERE id=$1`, id)
	return err
}

func (r *monitorRepo) CountByUserID(ctx context.Context, userID uuid.UUID) (int, error) {
	var count int
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM monitors WHERE user_id=$1`, userID).Scan(&count)
	return count, err
}

const monitorCols = `id, user_id, name, url, type, interval_seconds, timeout_seconds,
	failure_threshold, region, is_active, status, last_checked_at, created_at, updated_at`

func (r *monitorRepo) scanOne(row pgx.Row) (*domain.Monitor, error) {
	var m domain.Monitor
	err := row.Scan(
		&m.ID, &m.UserID, &m.Name, &m.URL, &m.Type, &m.IntervalSeconds, &m.TimeoutSeconds,
		&m.FailureThreshold, &m.Region, &m.IsActive, &m.Status, &m.LastCheckedAt, &m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("monitor scan: %w", err)
	}
	return &m, nil
}

func (r *monitorRepo) scanMany(rows pgx.Rows) ([]domain.Monitor, error) {
	defer rows.Close()
	var monitors []domain.Monitor
	for rows.Next() {
		m, err := r.scanOne(rows)
		if err != nil {
			return nil, err
		}
		monitors = append(monitors, *m)
	}
	return monitors, rows.Err()
}

// --- Check Repository ---

type checkRepo struct{ db *pgxpool.Pool }

func NewCheckRepository(db *pgxpool.Pool) outbound.CheckRepository {
	return &checkRepo{db: db}
}

func (r *checkRepo) Create(ctx context.Context, c *domain.MonitorCheck) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO monitor_checks (id, monitor_id, checked_at, is_up, status_code, response_time_ms, error_message, region)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		c.ID, c.MonitorID, c.CheckedAt, c.IsUp, c.StatusCode, c.ResponseTimeMs, c.ErrorMessage, c.Region,
	)
	return err
}

func (r *checkRepo) GetByMonitorID(ctx context.Context, monitorID uuid.UUID, from, to time.Time) ([]domain.MonitorCheck, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, monitor_id, checked_at, is_up, status_code, response_time_ms, error_message, region
		FROM monitor_checks
		WHERE monitor_id=$1 AND checked_at BETWEEN $2 AND $3
		ORDER BY checked_at ASC`, monitorID, from, to,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var checks []domain.MonitorCheck
	for rows.Next() {
		var c domain.MonitorCheck
		if err := rows.Scan(&c.ID, &c.MonitorID, &c.CheckedAt, &c.IsUp, &c.StatusCode, &c.ResponseTimeMs, &c.ErrorMessage, &c.Region); err != nil {
			return nil, err
		}
		checks = append(checks, c)
	}
	return checks, rows.Err()
}

func (r *checkRepo) GetUptimeStats(ctx context.Context, monitorID uuid.UUID, from time.Time) (float64, error) {
	var total, up int
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(*), COUNT(*) FILTER (WHERE is_up=true)
		FROM monitor_checks
		WHERE monitor_id=$1 AND checked_at >= $2`, monitorID, from,
	).Scan(&total, &up)
	if err != nil || total == 0 {
		return 100.0, err // no data = assume up
	}
	return float64(up) / float64(total) * 100, nil
}

func (r *checkRepo) GetLatest(ctx context.Context, monitorID uuid.UUID) (*domain.MonitorCheck, error) {
	var c domain.MonitorCheck
	err := r.db.QueryRow(ctx, `
		SELECT id, monitor_id, checked_at, is_up, status_code, response_time_ms, error_message, region
		FROM monitor_checks WHERE monitor_id=$1 ORDER BY checked_at DESC LIMIT 1`, monitorID,
	).Scan(&c.ID, &c.MonitorID, &c.CheckedAt, &c.IsUp, &c.StatusCode, &c.ResponseTimeMs, &c.ErrorMessage, &c.Region)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// --- Incident Repository ---

type incidentRepo struct{ db *pgxpool.Pool }

func NewIncidentRepository(db *pgxpool.Pool) outbound.IncidentRepository {
	return &incidentRepo{db: db}
}

func (r *incidentRepo) Create(ctx context.Context, i *domain.Incident) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO incidents (id, monitor_id, started_at) VALUES ($1,$2,$3)`,
		i.ID, i.MonitorID, i.StartedAt,
	)
	return err
}

func (r *incidentRepo) GetOpenByMonitorID(ctx context.Context, monitorID uuid.UUID) (*domain.Incident, error) {
	var i domain.Incident
	err := r.db.QueryRow(ctx,
		`SELECT id, monitor_id, started_at, resolved_at FROM incidents WHERE monitor_id=$1 AND resolved_at IS NULL`,
		monitorID,
	).Scan(&i.ID, &i.MonitorID, &i.StartedAt, &i.ResolvedAt)
	if err != nil {
		return nil, err
	}
	return &i, nil
}

func (r *incidentRepo) GetByMonitorID(ctx context.Context, monitorID uuid.UUID) ([]domain.Incident, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, monitor_id, started_at, resolved_at FROM incidents WHERE monitor_id=$1 ORDER BY started_at DESC LIMIT 50`,
		monitorID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var incidents []domain.Incident
	for rows.Next() {
		var i domain.Incident
		if err := rows.Scan(&i.ID, &i.MonitorID, &i.StartedAt, &i.ResolvedAt); err != nil {
			return nil, err
		}
		incidents = append(incidents, i)
	}
	return incidents, rows.Err()
}

func (r *incidentRepo) Resolve(ctx context.Context, incidentID uuid.UUID, resolvedAt time.Time) error {
	_, err := r.db.Exec(ctx,
		`UPDATE incidents SET resolved_at=$2 WHERE id=$1`,
		incidentID, resolvedAt,
	)
	return err
}
