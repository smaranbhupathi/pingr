// Package servicestest contains unit tests for the core service layer.
//
// Tests live in tests/services/ (separate from source) and only use
// exported interfaces — this is black-box testing. We verify the
// public contract, not internal implementation details.
package servicestest

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/smaranbhupathi/pingr/internal/core/domain"
)

var errNotFound = errors.New("not found")

// ── User repository ───────────────────────────────────────────────────────────

type mockUserRepo struct {
	users    map[uuid.UUID]*domain.User
	CreateFn func(ctx context.Context, user *domain.User) error
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{users: make(map[uuid.UUID]*domain.User)}
}

func (m *mockUserRepo) Create(ctx context.Context, user *domain.User) error {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, user)
	}
	m.users[user.ID] = user
	return nil
}

func (m *mockUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	u, ok := m.users[id]
	if !ok {
		return nil, errNotFound
	}
	return u, nil
}

func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	for _, u := range m.users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, errNotFound
}

func (m *mockUserRepo) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	for _, u := range m.users {
		if u.Username == username {
			return u, nil
		}
	}
	return nil, errNotFound
}

func (m *mockUserRepo) GetBySlug(ctx context.Context, slug string) (*domain.User, error) {
	for _, u := range m.users {
		if u.StatusPageSlug != nil && *u.StatusPageSlug == slug {
			return u, nil
		}
	}
	return nil, errNotFound
}

func (m *mockUserRepo) SetSlug(ctx context.Context, userID uuid.UUID, slug string) error {
	for _, u := range m.users {
		if u.ID == userID {
			u.StatusPageSlug = &slug
			return nil
		}
	}
	return errNotFound
}

func (m *mockUserRepo) GetByVerifyToken(ctx context.Context, token string) (*domain.User, error) {
	for _, u := range m.users {
		if u.VerifyToken == token {
			return u, nil
		}
	}
	return nil, errNotFound
}

func (m *mockUserRepo) GetByResetToken(ctx context.Context, token string) (*domain.User, error) {
	for _, u := range m.users {
		if u.ResetToken == token {
			return u, nil
		}
	}
	return nil, errNotFound
}

func (m *mockUserRepo) Update(ctx context.Context, user *domain.User) error {
	m.users[user.ID] = user
	return nil
}

// ── Plan repository ───────────────────────────────────────────────────────────

type mockPlanRepo struct {
	plans map[string]*domain.Plan
}

func newMockPlanRepo() *mockPlanRepo {
	freePlan := &domain.Plan{
		ID:                 uuid.New(),
		Name:               "free",
		MaxMonitors:        5,
		MinIntervalSeconds: 60,
	}
	proPlan := &domain.Plan{
		ID:                 uuid.New(),
		Name:               "pro",
		MaxMonitors:        50,
		MinIntervalSeconds: 30,
	}
	return &mockPlanRepo{
		plans: map[string]*domain.Plan{
			"free": freePlan,
			"pro":  proPlan,
		},
	}
}

func (m *mockPlanRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Plan, error) {
	for _, p := range m.plans {
		if p.ID == id {
			return p, nil
		}
	}
	return nil, errNotFound
}

func (m *mockPlanRepo) GetByName(ctx context.Context, name string) (*domain.Plan, error) {
	p, ok := m.plans[name]
	if !ok {
		return nil, errNotFound
	}
	return p, nil
}

// ── Monitor repository ────────────────────────────────────────────────────────

type mockMonitorRepo struct {
	monitors map[uuid.UUID]*domain.Monitor
}

func newMockMonitorRepo() *mockMonitorRepo {
	return &mockMonitorRepo{monitors: make(map[uuid.UUID]*domain.Monitor)}
}

func (m *mockMonitorRepo) Create(ctx context.Context, monitor *domain.Monitor) error {
	m.monitors[monitor.ID] = monitor
	return nil
}

func (m *mockMonitorRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Monitor, error) {
	mon, ok := m.monitors[id]
	if !ok {
		return nil, errNotFound
	}
	return mon, nil
}

func (m *mockMonitorRepo) GetByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Monitor, error) {
	var result []domain.Monitor
	for _, mon := range m.monitors {
		if mon.UserID == userID {
			result = append(result, *mon)
		}
	}
	return result, nil
}

func (m *mockMonitorRepo) GetByUsername(ctx context.Context, username string) ([]domain.Monitor, error) {
	return nil, nil
}

func (m *mockMonitorRepo) GetBySlug(ctx context.Context, slug string) ([]domain.Monitor, error) {
	return nil, nil
}

func (m *mockMonitorRepo) GetDue(ctx context.Context, region string) ([]domain.Monitor, error) {
	return nil, nil
}

func (m *mockMonitorRepo) Update(ctx context.Context, monitor *domain.Monitor) error {
	m.monitors[monitor.ID] = monitor
	return nil
}

func (m *mockMonitorRepo) Delete(ctx context.Context, id uuid.UUID) error {
	delete(m.monitors, id)
	return nil
}

func (m *mockMonitorRepo) CountByUserID(ctx context.Context, userID uuid.UUID) (int, error) {
	count := 0
	for _, mon := range m.monitors {
		if mon.UserID == userID {
			count++
		}
	}
	return count, nil
}

func (m *mockMonitorRepo) UpdateComponentStatus(ctx context.Context, id uuid.UUID, status domain.ComponentStatus) error {
	if mon, ok := m.monitors[id]; ok {
		mon.ComponentStatus = status
	}
	return nil
}

// ── Check repository ──────────────────────────────────────────────────────────

type mockCheckRepo struct{}

func (m *mockCheckRepo) Create(ctx context.Context, check *domain.MonitorCheck) error { return nil }
func (m *mockCheckRepo) GetByMonitorID(ctx context.Context, monitorID uuid.UUID, from, to time.Time) ([]domain.MonitorCheck, error) {
	return nil, nil
}
func (m *mockCheckRepo) GetUptimeStats(ctx context.Context, monitorID uuid.UUID, from time.Time) (float64, error) {
	return 100.0, nil
}
func (m *mockCheckRepo) GetDailyUptime(ctx context.Context, monitorID uuid.UUID, days int) ([]domain.DailyUptimeStat, error) {
	return nil, nil
}
func (m *mockCheckRepo) GetLatest(ctx context.Context, monitorID uuid.UUID) (*domain.MonitorCheck, error) {
	return nil, nil
}

// ── Incident repository ───────────────────────────────────────────────────────

type mockIncidentRepo struct{}

func (m *mockIncidentRepo) Create(ctx context.Context, incident *domain.Incident) error { return nil }
func (m *mockIncidentRepo) GetByID(ctx context.Context, id, userID uuid.UUID) (*domain.Incident, error) {
	return nil, errNotFound
}
func (m *mockIncidentRepo) ListByUser(ctx context.Context, userID uuid.UUID) ([]domain.Incident, error) {
	return nil, nil
}
func (m *mockIncidentRepo) ListByMonitor(ctx context.Context, monitorID uuid.UUID) ([]domain.Incident, error) {
	return nil, nil
}
func (m *mockIncidentRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.IncidentStatus, resolvedAt *time.Time) error {
	return nil
}
func (m *mockIncidentRepo) AddUpdate(ctx context.Context, update *domain.IncidentUpdate) error {
	return nil
}
func (m *mockIncidentRepo) GetUpdates(ctx context.Context, incidentID uuid.UUID) ([]domain.IncidentUpdate, error) {
	return nil, nil
}
func (m *mockIncidentRepo) GetOpenByMonitorID(ctx context.Context, monitorID uuid.UUID) (*domain.Incident, error) {
	return nil, errNotFound
}
func (m *mockIncidentRepo) GetByOutageEventID(ctx context.Context, outageEventID uuid.UUID) (*domain.Incident, error) {
	return nil, errNotFound
}
func (m *mockIncidentRepo) Resolve(ctx context.Context, incidentID uuid.UUID) error {
	return nil
}

// ── Component repository ──────────────────────────────────────────────────────

type mockComponentRepo struct {
	components map[uuid.UUID]*domain.Component
}

func newMockComponentRepo() *mockComponentRepo {
	return &mockComponentRepo{components: make(map[uuid.UUID]*domain.Component)}
}

func (m *mockComponentRepo) Create(ctx context.Context, c *domain.Component) error {
	m.components[c.ID] = c
	return nil
}

func (m *mockComponentRepo) GetByID(ctx context.Context, id, userID uuid.UUID) (*domain.Component, error) {
	c, ok := m.components[id]
	if !ok || c.UserID != userID {
		return nil, errNotFound
	}
	return c, nil
}

func (m *mockComponentRepo) ListByUser(ctx context.Context, userID uuid.UUID) ([]domain.Component, error) {
	var result []domain.Component
	for _, c := range m.components {
		if c.UserID == userID {
			result = append(result, *c)
		}
	}
	return result, nil
}

func (m *mockComponentRepo) Update(ctx context.Context, c *domain.Component) error {
	m.components[c.ID] = c
	return nil
}

func (m *mockComponentRepo) Delete(ctx context.Context, id, userID uuid.UUID) error {
	delete(m.components, id)
	return nil
}

// ── Alert channel repository ──────────────────────────────────────────────────

type mockAlertChannelRepo struct {
	channels map[uuid.UUID]*domain.AlertChannel
}

func newMockAlertChannelRepo() *mockAlertChannelRepo {
	return &mockAlertChannelRepo{channels: make(map[uuid.UUID]*domain.AlertChannel)}
}

func (m *mockAlertChannelRepo) Create(ctx context.Context, ch *domain.AlertChannel) error {
	m.channels[ch.ID] = ch
	return nil
}

func (m *mockAlertChannelRepo) GetByUserID(ctx context.Context, userID uuid.UUID) ([]domain.AlertChannel, error) {
	var result []domain.AlertChannel
	for _, ch := range m.channels {
		if ch.UserID == userID {
			result = append(result, *ch)
		}
	}
	return result, nil
}

func (m *mockAlertChannelRepo) GetByMonitorID(ctx context.Context, monitorID uuid.UUID) ([]domain.AlertChannel, error) {
	return nil, nil
}

func (m *mockAlertChannelRepo) GetByID(ctx context.Context, id, userID uuid.UUID) (*domain.AlertChannel, error) {
	ch, ok := m.channels[id]
	if !ok || ch.UserID != userID {
		return nil, errNotFound
	}
	return ch, nil
}

func (m *mockAlertChannelRepo) UpdateName(ctx context.Context, id, userID uuid.UUID, name string) error {
	ch, ok := m.channels[id]
	if !ok || ch.UserID != userID {
		return errNotFound
	}
	ch.Name = name
	return nil
}

func (m *mockAlertChannelRepo) UpdateEnabled(ctx context.Context, id, userID uuid.UUID, enabled bool) error {
	ch, ok := m.channels[id]
	if !ok || ch.UserID != userID {
		return errNotFound
	}
	ch.IsEnabled = enabled
	return nil
}

func (m *mockAlertChannelRepo) Delete(ctx context.Context, id uuid.UUID) error {
	delete(m.channels, id)
	return nil
}

// ── Alert subscription repository ────────────────────────────────────────────

type mockAlertSubRepo struct{}

func (m *mockAlertSubRepo) Create(ctx context.Context, sub *domain.AlertSubscription) error {
	return nil
}
func (m *mockAlertSubRepo) DeleteByMonitorAndChannel(ctx context.Context, monitorID, channelID uuid.UUID) error {
	return nil
}

// ── Email sender ──────────────────────────────────────────────────────────────

type mockEmailSender struct {
	verificationsSent []string
	resetsSent        []string
	confirmationsSent []string
}

func (m *mockEmailSender) SendVerification(ctx context.Context, toEmail, token string) error {
	m.verificationsSent = append(m.verificationsSent, toEmail)
	return nil
}

func (m *mockEmailSender) SendPasswordReset(ctx context.Context, toEmail, token string) error {
	m.resetsSent = append(m.resetsSent, toEmail)
	return nil
}

func (m *mockEmailSender) SendSubscriptionConfirmation(ctx context.Context, toEmail, monitorName, monitorURL string) error {
	m.confirmationsSent = append(m.confirmationsSent, toEmail)
	return nil
}
