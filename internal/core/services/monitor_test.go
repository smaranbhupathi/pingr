package services

// Monitor service tests.
//
// Covers the business rules that live in the service layer:
//   - Plan limits (max monitors per user)
//   - Minimum check interval enforcement
//   - Ownership checks (user can only touch their own monitors)
//   - Status transitions when pausing / resuming

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/smaranbhupathi/pingr/internal/core/domain"
	"github.com/smaranbhupathi/pingr/internal/core/ports/inbound"
)

// newTestMonitorService wires up a monitorService with all mocks.
func newTestMonitorService() (*monitorService, *mockUserRepo, *mockMonitorRepo, *mockPlanRepo) {
	userRepo := newMockUserRepo()
	monitorRepo := newMockMonitorRepo()
	planRepo := newMockPlanRepo()

	svc := NewMonitorService(
		monitorRepo,
		&mockCheckRepo{},
		&mockIncidentRepo{},
		userRepo,
		planRepo,
	).(*monitorService)

	return svc, userRepo, monitorRepo, planRepo
}

// seedUser adds a user with the given plan directly into the mock repos.
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
	ctx := context.Background()

	monitor, err := svc.Create(ctx, inbound.CreateMonitorInput{
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
	// Free plan: minimum interval is 60 seconds
	user := seedUser(userRepo, planRepo, "free")

	_, err := svc.Create(context.Background(), inbound.CreateMonitorInput{
		UserID:          user.ID,
		Name:            "Too Fast",
		URL:             "https://example.com",
		IntervalSeconds: 30, // below free plan minimum of 60
	})

	if !errors.Is(err, ErrInvalidInterval) {
		t.Errorf("expected ErrInvalidInterval, got %v", err)
	}
}

func TestCreate_ProPlanAllowsFasterInterval(t *testing.T) {
	svc, userRepo, _, planRepo := newTestMonitorService()
	// Pro plan: minimum interval is 30 seconds
	user := seedUser(userRepo, planRepo, "pro")

	_, err := svc.Create(context.Background(), inbound.CreateMonitorInput{
		UserID:          user.ID,
		Name:            "Fast Monitor",
		URL:             "https://example.com",
		IntervalSeconds: 30, // allowed on pro plan
	})

	if err != nil {
		t.Errorf("pro plan should allow 30s interval, got %v", err)
	}
}

func TestCreate_EnforcesMonitorLimit(t *testing.T) {
	svc, userRepo, monitorRepo, planRepo := newTestMonitorService()
	// Free plan: max 5 monitors
	user := seedUser(userRepo, planRepo, "free")
	ctx := context.Background()

	// Seed 5 monitors directly into the repo (already at the limit)
	for i := 0; i < 5; i++ {
		monitorRepo.monitors[uuid.New()] = &domain.Monitor{
			ID:     uuid.New(),
			UserID: user.ID,
		}
	}

	// 6th monitor should be rejected
	_, err := svc.Create(ctx, inbound.CreateMonitorInput{
		UserID:          user.ID,
		Name:            "One Too Many",
		URL:             "https://example.com",
		IntervalSeconds: 60,
	})

	if !errors.Is(err, ErrMonitorLimitReached) {
		t.Errorf("expected ErrMonitorLimitReached, got %v", err)
	}
}

// ── Update ────────────────────────────────────────────────────────────────────

func TestUpdate_PauseMonitor(t *testing.T) {
	svc, userRepo, monitorRepo, planRepo := newTestMonitorService()
	user := seedUser(userRepo, planRepo, "free")
	ctx := context.Background()

	// Create a monitor first
	monitor, _ := svc.Create(ctx, inbound.CreateMonitorInput{
		UserID:          user.ID,
		Name:            "My API",
		URL:             "https://example.com",
		IntervalSeconds: 60,
	})

	// Manually set it to "up" so we can verify pause changes the status
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
	// Status must change to 'paused' — not remain 'up'
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

	// Manually pause it
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
	// Status resets to 'pending' — worker will re-check on next tick
	if updated.Status != domain.MonitorStatusPending {
		t.Errorf("resumed monitor should have status 'pending', got %q", updated.Status)
	}
}

func TestUpdate_RejectsIntervalBelowPlanMinimum(t *testing.T) {
	svc, userRepo, _, planRepo := newTestMonitorService()
	user := seedUser(userRepo, planRepo, "free") // free: min 60s
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

	if !errors.Is(err, ErrInvalidInterval) {
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

	otherUserID := uuid.New() // a completely different user
	newName := "Hacked"
	_, err := svc.Update(ctx, monitor.ID, otherUserID, inbound.UpdateMonitorInput{
		Name: &newName,
	})

	if !errors.Is(err, ErrMonitorNotFound) {
		// Returns "not found" intentionally — doesn't reveal the monitor exists
		t.Errorf("expected ErrMonitorNotFound when wrong user tries to update, got %v", err)
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

	// Monitor should be gone from the repo
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

	otherUserID := uuid.New()
	err := svc.Delete(ctx, monitor.ID, otherUserID)

	if !errors.Is(err, ErrMonitorNotFound) {
		t.Errorf("expected ErrMonitorNotFound when wrong user deletes, got %v", err)
	}
}
