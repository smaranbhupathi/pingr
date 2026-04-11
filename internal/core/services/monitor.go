package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/smaranbhupathi/pingr/internal/core/domain"
	"github.com/smaranbhupathi/pingr/internal/core/ports/inbound"
	"github.com/smaranbhupathi/pingr/internal/core/ports/outbound"
)

var (
	ErrMonitorNotFound   = errors.New("monitor not found")
	ErrMonitorLimitReached = errors.New("monitor limit reached for your plan")
	ErrInvalidInterval   = errors.New("check interval below plan minimum")
)

const defaultRegion = "sin" // Singapore — matches Railway worker region

type monitorService struct {
	monitors     outbound.MonitorRepository
	checks       outbound.CheckRepository
	incidents    outbound.IncidentRepository
	users        outbound.UserRepository
	plans        outbound.PlanRepository
}

func NewMonitorService(
	monitors outbound.MonitorRepository,
	checks outbound.CheckRepository,
	incidents outbound.IncidentRepository,
	users outbound.UserRepository,
	plans outbound.PlanRepository,
) inbound.MonitorService {
	return &monitorService{
		monitors:  monitors,
		checks:    checks,
		incidents: incidents,
		users:     users,
		plans:     plans,
	}
}

func (s *monitorService) Create(ctx context.Context, input inbound.CreateMonitorInput) (*domain.Monitor, error) {
	user, err := s.users.GetByID(ctx, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}

	plan, err := s.plans.GetByID(ctx, user.PlanID)
	if err != nil {
		return nil, fmt.Errorf("get plan: %w", err)
	}

	count, err := s.monitors.CountByUserID(ctx, input.UserID)
	if err != nil {
		return nil, err
	}
	if count >= plan.MaxMonitors {
		return nil, ErrMonitorLimitReached
	}

	if input.IntervalSeconds < plan.MinIntervalSeconds {
		return nil, ErrInvalidInterval
	}

	monitor := &domain.Monitor{
		ID:               uuid.New(),
		UserID:           input.UserID,
		Name:             input.Name,
		URL:              input.URL,
		Type:             domain.MonitorTypeHTTP,
		IntervalSeconds:  input.IntervalSeconds,
		TimeoutSeconds:   30,
		FailureThreshold: 2, // configurable later via plan or per-monitor setting
		Region:           defaultRegion,
		IsActive:         true,
		Status:           domain.MonitorStatusPending,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if err := s.monitors.Create(ctx, monitor); err != nil {
		return nil, fmt.Errorf("create monitor: %w", err)
	}

	return monitor, nil
}

func (s *monitorService) GetByID(ctx context.Context, id, userID uuid.UUID) (*inbound.MonitorDetail, error) {
	monitor, err := s.monitors.GetByID(ctx, id)
	if err != nil || monitor.UserID != userID {
		return nil, ErrMonitorNotFound
	}

	return s.buildDetail(ctx, *monitor)
}

func (s *monitorService) ListByUser(ctx context.Context, userID uuid.UUID) ([]domain.Monitor, error) {
	return s.monitors.GetByUserID(ctx, userID)
}

func (s *monitorService) Update(ctx context.Context, id, userID uuid.UUID, input inbound.UpdateMonitorInput) (*domain.Monitor, error) {
	monitor, err := s.monitors.GetByID(ctx, id)
	if err != nil || monitor.UserID != userID {
		return nil, ErrMonitorNotFound
	}

	if input.Name != nil {
		monitor.Name = *input.Name
	}
	if input.IntervalSeconds != nil {
		user, _ := s.users.GetByID(ctx, userID)
		plan, _ := s.plans.GetByID(ctx, user.PlanID)
		if *input.IntervalSeconds < plan.MinIntervalSeconds {
			return nil, ErrInvalidInterval
		}
		monitor.IntervalSeconds = *input.IntervalSeconds
	}
	if input.IsActive != nil {
		monitor.IsActive = *input.IsActive
		if !*input.IsActive {
			monitor.Status = domain.MonitorStatusPaused
		} else {
			monitor.Status = domain.MonitorStatusPending
		}
	}

	monitor.UpdatedAt = time.Now()

	if err := s.monitors.Update(ctx, monitor); err != nil {
		return nil, err
	}

	return monitor, nil
}

func (s *monitorService) Delete(ctx context.Context, id, userID uuid.UUID) error {
	monitor, err := s.monitors.GetByID(ctx, id)
	if err != nil || monitor.UserID != userID {
		return ErrMonitorNotFound
	}
	return s.monitors.Delete(ctx, id)
}

func (s *monitorService) GetResponseTimeGraph(ctx context.Context, monitorID uuid.UUID, from, to time.Time) ([]inbound.CheckDataPoint, error) {
	checks, err := s.checks.GetByMonitorID(ctx, monitorID, from, to)
	if err != nil {
		return nil, err
	}

	points := make([]inbound.CheckDataPoint, len(checks))
	for i, c := range checks {
		points[i] = inbound.CheckDataPoint{
			Timestamp:      c.CheckedAt,
			ResponseTimeMs: c.ResponseTimeMs,
			IsUp:           c.IsUp,
		}
	}
	return points, nil
}

func (s *monitorService) GetStatusPage(ctx context.Context, username string) ([]inbound.MonitorDetail, error) {
	monitors, err := s.monitors.GetByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	details := make([]inbound.MonitorDetail, 0, len(monitors))
	for _, m := range monitors {
		detail, err := s.buildDetail(ctx, m)
		if err != nil {
			continue
		}
		details = append(details, *detail)
	}

	return details, nil
}

func (s *monitorService) buildDetail(ctx context.Context, monitor domain.Monitor) (*inbound.MonitorDetail, error) {
	now := time.Now()

	u24h, _ := s.checks.GetUptimeStats(ctx, monitor.ID, now.Add(-24*time.Hour))
	u7d, _ := s.checks.GetUptimeStats(ctx, monitor.ID, now.Add(-7*24*time.Hour))
	u30d, _ := s.checks.GetUptimeStats(ctx, monitor.ID, now.Add(-30*24*time.Hour))
	u90d, _ := s.checks.GetUptimeStats(ctx, monitor.ID, now.Add(-90*24*time.Hour))

	daily, _ := s.checks.GetDailyUptime(ctx, monitor.ID, 90)
	latest, _ := s.checks.GetLatest(ctx, monitor.ID)
	incidents, _ := s.incidents.ListByMonitor(ctx, monitor.ID)

	var activeIncident *domain.Incident
	openInc, err := s.incidents.GetOpenByMonitorID(ctx, monitor.ID)
	if err == nil && openInc != nil {
		activeIncident = openInc
	}

	return &inbound.MonitorDetail{
		Monitor: monitor,
		Uptime: inbound.UptimeStats{
			Last24h: u24h,
			Last7d:  u7d,
			Last30d: u30d,
			Last90d: u90d,
		},
		DailyUptime:    daily,
		RecentCheck:    latest,
		Incidents:      incidents,
		ActiveIncident: activeIncident,
	}, nil
}
