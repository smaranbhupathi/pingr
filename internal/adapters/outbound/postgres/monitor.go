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
			(id, user_id, name, description, url, type, interval_seconds, timeout_seconds,
			 failure_threshold, region, is_active, status, component_status, component_id, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)`,
		m.ID, m.UserID, m.Name, m.Description, m.URL, m.Type, m.IntervalSeconds, m.TimeoutSeconds,
		m.FailureThreshold, m.Region, m.IsActive, m.Status, m.ComponentStatus, m.ComponentID, m.CreatedAt, m.UpdatedAt,
	)
	return err
}

func (r *monitorRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Monitor, error) {
	return r.scanOne(r.db.QueryRow(ctx, `SELECT `+monitorCols+` FROM monitors m WHERE m.id=$1 AND m.deleted_at IS NULL`, id))
}

func (r *monitorRepo) GetByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Monitor, error) {
	rows, err := r.db.Query(ctx, `SELECT `+monitorCols+` FROM monitors m WHERE m.user_id=$1 AND m.deleted_at IS NULL ORDER BY m.created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	return r.scanMany(rows)
}

func (r *monitorRepo) GetByUsername(ctx context.Context, username string) ([]domain.Monitor, error) {
	rows, err := r.db.Query(ctx, `
		SELECT `+monitorCols+`
		FROM monitors m
		JOIN users u ON u.id = m.user_id
		WHERE u.username=$1 AND m.is_active=true AND m.deleted_at IS NULL
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
		SELECT `+monitorCols+` FROM monitors m
		WHERE m.is_active = true
		  AND m.deleted_at IS NULL
		  AND m.region = $1
		  AND (
		    m.last_checked_at IS NULL
		    OR m.last_checked_at <= NOW() - (m.interval_seconds || ' seconds')::interval
		  )
		ORDER BY m.last_checked_at ASC NULLS FIRST
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
			name=$2, description=$3, url=$4, interval_seconds=$5, timeout_seconds=$6,
			failure_threshold=$7, is_active=$8, status=$9, component_status=$10,
			component_id=$11, last_checked_at=$12, updated_at=$13
		WHERE id=$1`,
		m.ID, m.Name, m.Description, m.URL, m.IntervalSeconds, m.TimeoutSeconds,
		m.FailureThreshold, m.IsActive, m.Status, m.ComponentStatus,
		m.ComponentID, m.LastCheckedAt, m.UpdatedAt,
	)
	return err
}

func (r *monitorRepo) UpdateComponentStatus(ctx context.Context, id uuid.UUID, status domain.ComponentStatus) error {
	_, err := r.db.Exec(ctx,
		`UPDATE monitors SET component_status=$2, updated_at=NOW() WHERE id=$1 AND deleted_at IS NULL`,
		id, status,
	)
	return err
}

func (r *monitorRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `UPDATE monitors SET deleted_at = NOW() WHERE id=$1 AND deleted_at IS NULL`, id)
	return err
}

func (r *monitorRepo) CountByUserID(ctx context.Context, userID uuid.UUID) (int, error) {
	var count int
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM monitors WHERE user_id=$1 AND deleted_at IS NULL`, userID).Scan(&count)
	return count, err
}

const monitorCols = `m.id, m.user_id, m.name, m.description, m.url, m.type, m.interval_seconds, m.timeout_seconds,
	m.failure_threshold, m.region, m.is_active, m.status, m.component_status, m.component_id, m.last_checked_at, m.created_at, m.updated_at`

func (r *monitorRepo) scanOne(row pgx.Row) (*domain.Monitor, error) {
	var m domain.Monitor
	err := row.Scan(
		&m.ID, &m.UserID, &m.Name, &m.Description, &m.URL, &m.Type, &m.IntervalSeconds, &m.TimeoutSeconds,
		&m.FailureThreshold, &m.Region, &m.IsActive, &m.Status, &m.ComponentStatus, &m.ComponentID,
		&m.LastCheckedAt, &m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("monitor scan: %w", err)
	}
	return &m, nil
}

func (r *monitorRepo) scanMany(rows pgx.Rows) ([]domain.Monitor, error) {
	defer rows.Close()
	monitors := make([]domain.Monitor, 0)
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

	checks := make([]domain.MonitorCheck, 0)
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

func (r *checkRepo) GetDailyUptime(ctx context.Context, monitorID uuid.UUID, days int) ([]domain.DailyUptimeStat, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			(checked_at AT TIME ZONE 'UTC')::date AS day,
			ROUND(
				COUNT(*) FILTER (WHERE is_up = true)::numeric /
				NULLIF(COUNT(*)::numeric, 0) * 100,
			2) AS uptime_pct
		FROM monitor_checks
		WHERE monitor_id = $1
		  AND checked_at >= NOW() - ($2::int || ' days')::interval
		GROUP BY day
		ORDER BY day ASC`, monitorID, days,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Build a map of day → uptime%
	type row struct {
		date   string
		uptime float64
	}
	dataByDate := make(map[string]float64)
	for rows.Next() {
		var day string
		var pct float64
		if err := rows.Scan(&day, &pct); err != nil {
			return nil, err
		}
		dataByDate[day] = pct
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Fill the full window, marking missing days as -1 (no data)
	result := make([]domain.DailyUptimeStat, days)
	for i := 0; i < days; i++ {
		d := time.Now().UTC().AddDate(0, 0, -(days-1-i)).Format("2006-01-02")
		uptime, ok := dataByDate[d]
		if !ok {
			uptime = -1
		}
		result[i] = domain.DailyUptimeStat{Date: d, Uptime: uptime}
	}
	return result, nil
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

// --- OutageEvent Repository (worker-internal, used for uptime math) ---

type outageEventRepo struct{ db *pgxpool.Pool }

func NewOutageEventRepository(db *pgxpool.Pool) outbound.OutageEventRepository {
	return &outageEventRepo{db: db}
}

func (r *outageEventRepo) Create(ctx context.Context, e *domain.OutageEvent) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO outage_events (id, monitor_id, started_at) VALUES ($1,$2,$3)`,
		e.ID, e.MonitorID, e.StartedAt,
	)
	return err
}

func (r *outageEventRepo) GetOpenByMonitorID(ctx context.Context, monitorID uuid.UUID) (*domain.OutageEvent, error) {
	var e domain.OutageEvent
	err := r.db.QueryRow(ctx,
		`SELECT id, monitor_id, started_at, resolved_at FROM outage_events WHERE monitor_id=$1 AND resolved_at IS NULL`,
		monitorID,
	).Scan(&e.ID, &e.MonitorID, &e.StartedAt, &e.ResolvedAt)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *outageEventRepo) Resolve(ctx context.Context, eventID uuid.UUID, resolvedAt time.Time) error {
	_, err := r.db.Exec(ctx,
		`UPDATE outage_events SET resolved_at=$2 WHERE id=$1`,
		eventID, resolvedAt,
	)
	return err
}

// --- Incident Repository (user-facing, shown on status page) ---

type incidentRepo struct{ db *pgxpool.Pool }

func NewIncidentRepository(db *pgxpool.Pool) outbound.IncidentRepository {
	return &incidentRepo{db: db}
}

func (r *incidentRepo) Create(ctx context.Context, inc *domain.Incident) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx,
		`INSERT INTO incidents (id, user_id, name, status, source, outage_event_id, resolved_at, created_at, updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		inc.ID, inc.UserID, inc.Name, inc.Status, inc.Source, inc.OutageEventID, inc.ResolvedAt, inc.CreatedAt, inc.UpdatedAt,
	)
	if err != nil {
		return err
	}

	for _, mID := range inc.MonitorIDs {
		_, err = tx.Exec(ctx,
			`INSERT INTO incident_affected_monitors (incident_id, monitor_id) VALUES ($1,$2)`,
			inc.ID, mID,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *incidentRepo) GetByID(ctx context.Context, id, userID uuid.UUID) (*domain.Incident, error) {
	inc, err := r.scanOne(r.db.QueryRow(ctx,
		`SELECT id, user_id, name, status, source, outage_event_id, resolved_at, created_at, updated_at
		 FROM incidents WHERE id=$1 AND user_id=$2`, id, userID,
	))
	if err != nil {
		return nil, err
	}
	if err := r.attachDetails(ctx, inc); err != nil {
		return nil, err
	}
	return inc, nil
}

func (r *incidentRepo) ListByUser(ctx context.Context, userID uuid.UUID) ([]domain.Incident, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, user_id, name, status, source, outage_event_id, resolved_at, created_at, updated_at
		 FROM incidents WHERE user_id=$1 ORDER BY created_at DESC`, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.scanMany(ctx, rows)
}

func (r *incidentRepo) ListByMonitor(ctx context.Context, monitorID uuid.UUID) ([]domain.Incident, error) {
	rows, err := r.db.Query(ctx,
		`SELECT i.id, i.user_id, i.name, i.status, i.source, i.outage_event_id, i.resolved_at, i.created_at, i.updated_at
		 FROM incidents i
		 JOIN incident_affected_monitors iam ON iam.incident_id = i.id
		 WHERE iam.monitor_id=$1
		 ORDER BY i.created_at DESC LIMIT 20`, monitorID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.scanMany(ctx, rows)
}

func (r *incidentRepo) GetByOutageEventID(ctx context.Context, outageEventID uuid.UUID) (*domain.Incident, error) {
	return r.scanOne(r.db.QueryRow(ctx,
		`SELECT id, user_id, name, status, source, outage_event_id, resolved_at, created_at, updated_at
		 FROM incidents WHERE outage_event_id=$1`, outageEventID,
	))
}

func (r *incidentRepo) GetOpenByMonitorID(ctx context.Context, monitorID uuid.UUID) (*domain.Incident, error) {
	inc, err := r.scanOne(r.db.QueryRow(ctx,
		`SELECT i.id, i.user_id, i.name, i.status, i.source, i.resolved_at, i.created_at, i.updated_at
		 FROM incidents i
		 JOIN incident_affected_monitors iam ON iam.incident_id = i.id
		 WHERE iam.monitor_id=$1 AND i.resolved_at IS NULL
		 ORDER BY i.created_at DESC LIMIT 1`, monitorID,
	))
	if err != nil {
		return nil, err
	}
	return inc, nil
}

func (r *incidentRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.IncidentStatus, resolvedAt *time.Time) error {
	_, err := r.db.Exec(ctx,
		`UPDATE incidents SET status=$2, resolved_at=$3, updated_at=NOW() WHERE id=$1`,
		id, status, resolvedAt,
	)
	return err
}

func (r *incidentRepo) AddUpdate(ctx context.Context, u *domain.IncidentUpdate) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO incident_updates (id, incident_id, status, message, notify, source, created_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		u.ID, u.IncidentID, u.Status, u.Message, u.Notify, u.Source, u.CreatedAt,
	)
	return err
}

func (r *incidentRepo) GetUpdates(ctx context.Context, incidentID uuid.UUID) ([]domain.IncidentUpdate, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, incident_id, status, message, notify, source, created_at
		 FROM incident_updates WHERE incident_id=$1 ORDER BY created_at ASC`, incidentID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	updates := make([]domain.IncidentUpdate, 0)
	for rows.Next() {
		var u domain.IncidentUpdate
		if err := rows.Scan(&u.ID, &u.IncidentID, &u.Status, &u.Message, &u.Notify, &u.Source, &u.CreatedAt); err != nil {
			return nil, err
		}
		updates = append(updates, u)
	}
	return updates, rows.Err()
}

func (r *incidentRepo) Resolve(ctx context.Context, incidentID uuid.UUID) error {
	now := time.Now()
	_, err := r.db.Exec(ctx,
		`UPDATE incidents SET status='resolved', resolved_at=$2, updated_at=$2 WHERE id=$1`,
		incidentID, now,
	)
	return err
}

// scanOne scans a single incident row (no updates or monitors attached yet).
func (r *incidentRepo) scanOne(row interface {
	Scan(...any) error
}) (*domain.Incident, error) {
	var inc domain.Incident
	if err := row.Scan(&inc.ID, &inc.UserID, &inc.Name, &inc.Status, &inc.Source, &inc.OutageEventID, &inc.ResolvedAt, &inc.CreatedAt, &inc.UpdatedAt); err != nil {
		return nil, err
	}
	return &inc, nil
}

func (r *incidentRepo) scanMany(ctx context.Context, rows interface {
	Next() bool
	Scan(...any) error
	Err() error
}) ([]domain.Incident, error) {
	incidents := make([]domain.Incident, 0)
	for rows.Next() {
		var inc domain.Incident
		if err := rows.Scan(&inc.ID, &inc.UserID, &inc.Name, &inc.Status, &inc.Source, &inc.OutageEventID, &inc.ResolvedAt, &inc.CreatedAt, &inc.UpdatedAt); err != nil {
			return nil, err
		}
		incidents = append(incidents, inc)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	// Attach updates + monitor IDs for each
	for i := range incidents {
		if err := r.attachDetails(ctx, &incidents[i]); err != nil {
			return nil, err
		}
	}
	return incidents, nil
}

func (r *incidentRepo) attachDetails(ctx context.Context, inc *domain.Incident) error {
	updates, err := r.GetUpdates(ctx, inc.ID)
	if err != nil {
		return err
	}
	inc.Updates = updates

	rows, err := r.db.Query(ctx,
		`SELECT m.id, m.name, m.url
		 FROM incident_affected_monitors iam
		 JOIN monitors m ON m.id = iam.monitor_id
		 WHERE iam.incident_id=$1`, inc.ID,
	)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var m domain.IncidentMonitor
		if err := rows.Scan(&m.ID, &m.Name, &m.URL); err != nil {
			return err
		}
		inc.MonitorIDs = append(inc.MonitorIDs, m.ID)
		inc.Monitors = append(inc.Monitors, m)
	}
	return rows.Err()
}
