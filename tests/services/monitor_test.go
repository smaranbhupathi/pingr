package servicestest

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/smaranbhupathi/pingr/internal/core/domain"
	"github.com/smaranbhupathi/pingr/internal/core/ports/inbound"
	"github.com/smaranbhupathi/pingr/internal/core/services"
)

func newTestMonitorService() (inbound.MonitorService, *mockUserRepo, *mockMonitorRepo, *mockPlanRepo) {
	userRepo := newMockUserRepo()
	monitorRepo := newMockMonitorRepo()
	planRepo := newMockPlanRepo()

	svc := services.NewMonitorService(
		monitorRepo,
		&mockCheckRepo{},
		&mockIncidentRepo{},
		userRepo,
		planRepo,
	)

	return svc, userRepo, monitorRepo, planRepo
}

// seedUser adds a user with the given plan into the mock repos.
// Must use the same planRepo instance that was passed to the service.
func seedUser(userRepo *mockUserRepo, planRepo *mockPlanRepo, planName string) *domain.User {
	plan, _ := planRepo.GetByName(context.Background(), planName)
	user := &domain.User{
		ID:     uuid.New(),
		Email:  "user@example.com",
		PlanID: plan.ID,
	}
	userRepo.users[user.ID] = user
	return user
}

// ── Create ────────────────────────────────────────────────────────────────────

func TestCreate_Success(t *testing.T) {
	svc, userRepo, _, planRepo := newTestMonitorService()
	user := seedUser(userRepo, planRepo, "free")

	monitor, err := svc.Create(context.Background(), inbound.CreateMonitorInput{
		UserID:          user.ID,
		Name:            "My API",
		URL:             "https://example.com/health",
		IntervalSeconds: 60,
	})

	if err != nil {
		t.Fatalf("expected monitor to be created, got %v", err)
	}
	if monitor.ID == uuid.Nil {
		t.Error("monitor should have an ID assigned")
	}
	if monitor.Status != domain.MonitorStatusPending {
		t.Errorf("new monitor should have status 'pending', got %q", monitor.Status)
	}
	if !monitor.IsActive {
		t.Error("new monitor should be active by default")
	}
}

func TestCreate_RejectsIntervalBelowPlanMinimum(t *testing.T) {
	svc, userRepo, _, planRepo := newTestMonitorService()
	user := seedUser(userRepo, planRepo, "free") // free plan: min 60s

	_, err := svc.Create(context.Background(), inbound.CreateMonitorInput{
		UserID:          user.ID,
		Name:            "Too Fast",
		URL:             "https://example.com",
		IntervalSeconds: 30, // below free plan minimum
	})

	if !errors.Is(err, services.ErrInvalidInterval) {
		t.Errorf("expected ErrInvalidInterval, got %v", err)
	}
}

func TestCreate_ProPlanAllowsFasterInterval(t *testing.T) {
	svc, userRepo, _, planRepo := newTestMonitorService()
	user := seedUser(userRepo, planRepo, "pro") // pro plan: min 30s

	_, err := svc.Create(context.Background(), inbound.CreateMonitorInput{
		UserID:          user.ID,
		Name:            "Fast Monitor",
		URL:             "https://example.com",
		IntervalSeconds: 30,
	})

	if err != nil {
		t.Errorf("pro plan should allow 30s interval, got %v", err)
	}
}

func TestCreate_EnforcesMonitorLimit(t *testing.T) {
	svc, userRepo, monitorRepo, planRepo := newTestMonitorService()
	user := seedUser(userRepo, planRepo, "free") // free plan: max 5 monitors

	// Seed 5 monitors directly — already at the limit
	for i := 0; i < 5; i++ {
		id := uuid.New()
		monitorRepo.monitors[id] = &domain.Monitor{ID: id, UserID: user.ID}
	}

	_, err := svc.Create(context.Background(), inbound.CreateMonitorInput{
		UserID:          user.ID,
		Name:            "One Too Many",
		URL:             "https://example.com",
		IntervalSeconds: 60,
	})

	if !errors.Is(err, services.ErrMonitorLimitReached) {
		t.Errorf("expected ErrMonitorLimitReached, got %v", err)
	}
}

// ── Update ────────────────────────────────────────────────────────────────────

func TestUpdate_PauseMonitor(t *testing.T) {
	svc, userRepo, monitorRepo, planRepo := newTestMonitorService()
	user := seedUser(userRepo, planRepo, "free")
	ctx := context.Background()

	monitor, _ := svc.Create(ctx, inbound.CreateMonitorInput{
		UserID:          user.ID,
		Name:            "My API",
		URL:             "https://example.com",
		IntervalSeconds: 60,
	})
	monitorRepo.monitors[monitor.ID].Status = domain.MonitorStatusUp

	inactive := false
	updated, err := svc.Update(ctx, monitor.ID, user.ID, inbound.UpdateMonitorInput{
		IsActive: &inactive,
	})

	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if updated.IsActive {
		t.Error("monitor should be inactive after pausing")
	}
	if updated.Status != domain.MonitorStatusPaused {
		t.Errorf("paused monitor should have status 'paused', got %q", updated.Status)
	}
}

func TestUpdate_ResumeMonitor(t *testing.T) {
	svc, userRepo, monitorRepo, planRepo := newTestMonitorService()
	user := seedUser(userRepo, planRepo, "free")
	ctx := context.Background()

	monitor, _ := svc.Create(ctx, inbound.CreateMonitorInput{
		UserID:          user.ID,
		Name:            "My API",
		URL:             "https://example.com",
		IntervalSeconds: 60,
	})
	monitorRepo.monitors[monitor.ID].IsActive = false
	monitorRepo.monitors[monitor.ID].Status = domain.MonitorStatusPaused

	active := true
	updated, err := svc.Update(ctx, monitor.ID, user.ID, inbound.UpdateMonitorInput{
		IsActive: &active,
	})

	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if !updated.IsActive {
		t.Error("monitor should be active after resuming")
	}
	if updated.Status != domain.MonitorStatusPending {
		t.Errorf("resumed monitor should have status 'pending', got %q", updated.Status)
	}
}

func TestUpdate_RejectsIntervalBelowPlanMinimum(t *testing.T) {
	svc, userRepo, _, planRepo := newTestMonitorService()
	user := seedUser(userRepo, planRepo, "free")
	ctx := context.Background()

	monitor, _ := svc.Create(ctx, inbound.CreateMonitorInput{
		UserID:          user.ID,
		Name:            "My API",
		URL:             "https://example.com",
		IntervalSeconds: 60,
	})

	tooFast := 30
	_, err := svc.Update(ctx, monitor.ID, user.ID, inbound.UpdateMonitorInput{
		IntervalSeconds: &tooFast,
	})

	if !errors.Is(err, services.ErrInvalidInterval) {
		t.Errorf("expected ErrInvalidInterval, got %v", err)
	}
}

func TestUpdate_RejectsWrongOwner(t *testing.T) {
	svc, userRepo, _, planRepo := newTestMonitorService()
	user := seedUser(userRepo, planRepo, "free")
	ctx := context.Background()

	monitor, _ := svc.Create(ctx, inbound.CreateMonitorInput{
		UserID:          user.ID,
		Name:            "My API",
		URL:             "https://example.com",
		IntervalSeconds: 60,
	})

	otherUserID := uuid.New()
	newName := "Hacked"
	_, err := svc.Update(ctx, monitor.ID, otherUserID, inbound.UpdateMonitorInput{
		Name: &newName,
	})

	if !errors.Is(err, services.ErrMonitorNotFound) {
		t.Errorf("expected ErrMonitorNotFound when wrong user updates, got %v", err)
	}
}

// ── Delete ────────────────────────────────────────────────────────────────────

func TestDelete_Success(t *testing.T) {
	svc, userRepo, monitorRepo, planRepo := newTestMonitorService()
	user := seedUser(userRepo, planRepo, "free")
	ctx := context.Background()

	monitor, _ := svc.Create(ctx, inbound.CreateMonitorInput{
		UserID:          user.ID,
		Name:            "To Delete",
		URL:             "https://example.com",
		IntervalSeconds: 60,
	})

	if err := svc.Delete(ctx, monitor.ID, user.ID); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	if _, exists := monitorRepo.monitors[monitor.ID]; exists {
		t.Error("monitor should be removed after deletion")
	}
}

func TestDelete_RejectsWrongOwner(t *testing.T) {
	svc, userRepo, _, planRepo := newTestMonitorService()
	user := seedUser(userRepo, planRepo, "free")
	ctx := context.Background()

	monitor, _ := svc.Create(ctx, inbound.CreateMonitorInput{
		UserID:          user.ID,
		Name:            "Mine",
		URL:             "https://example.com",
		IntervalSeconds: 60,
	})

	err := svc.Delete(ctx, monitor.ID, uuid.New())
	if !errors.Is(err, services.ErrMonitorNotFound) {
		t.Errorf("expected ErrMonitorNotFound when wrong user deletes, got %v", err)
	}
}
