package services

// Auth service tests.
//
// Each test function is named Test<Method>_<scenario>.
// This makes it easy to see at a glance what's covered:
//
//   go test ./internal/core/services/... -v
//
// No database, no network — all mocks run in memory.

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/smaranbhupathi/pingr/internal/core/ports/inbound"
)

// newTestAuthService wires up an authService with all mocks.
// Each test gets a fresh instance so there's no shared state between tests.
func newTestAuthService() (*authService, *mockUserRepo, *mockEmailSender) {
	userRepo := newMockUserRepo()
	planRepo := newMockPlanRepo()
	channelRepo := newMockAlertChannelRepo()
	emailSender := &mockEmailSender{}

	svc := NewAuthService(userRepo, planRepo, channelRepo, emailSender, AuthServiceConfig{
		JWTSecret:            "test-secret",
		AccessTokenDuration:  15 * time.Minute,
		RefreshTokenDuration: 7 * 24 * time.Hour,
		AppBaseURL:           "http://localhost",
	}).(*authService)

	return svc, userRepo, emailSender
}

// ── Register ──────────────────────────────────────────────────────────────────

func TestRegister_Success(t *testing.T) {
	svc, userRepo, emailSender := newTestAuthService()
	ctx := context.Background()

	err := svc.Register(ctx, inbound.RegisterInput{
		Email:    "alice@example.com",
		Username: "alice",
		Password: "password123",
	})

	// No error
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// User was stored
	user, err := userRepo.GetByEmail(ctx, "alice@example.com")
	if err != nil {
		t.Fatal("user should exist in repo after registration")
	}

	// Password should be hashed, not stored in plain text
	if user.PasswordHash == "password123" {
		t.Error("password must not be stored in plain text")
	}

	// User should not be verified yet — they need to click the email link
	if user.IsVerified {
		t.Error("new user should not be verified until email is confirmed")
	}

	// Verification token must be set
	if user.VerifyToken == "" {
		t.Error("verify token should be set after registration")
	}

	// Verification email should have been sent
	if len(emailSender.verificationsSent) != 1 || emailSender.verificationsSent[0] != "alice@example.com" {
		t.Error("verification email should be sent to the registered address")
	}
}

func TestRegister_DuplicateEmail(t *testing.T) {
	svc, _, _ := newTestAuthService()
	ctx := context.Background()

	input := inbound.RegisterInput{
		Email:    "alice@example.com",
		Username: "alice",
		Password: "password123",
	}

	// First registration — should succeed
	if err := svc.Register(ctx, input); err != nil {
		t.Fatalf("first registration failed: %v", err)
	}

	// Second registration with same email — should fail
	input.Username = "alice2" // different username, same email
	err := svc.Register(ctx, input)

	if !errors.Is(err, ErrEmailTaken) {
		t.Errorf("expected ErrEmailTaken, got %v", err)
	}
}

func TestRegister_DuplicateUsername(t *testing.T) {
	svc, _, _ := newTestAuthService()
	ctx := context.Background()

	// First registration
	if err := svc.Register(ctx, inbound.RegisterInput{
		Email:    "alice@example.com",
		Username: "alice",
		Password: "password123",
	}); err != nil {
		t.Fatalf("first registration failed: %v", err)
	}

	// Second registration — same username, different email
	err := svc.Register(ctx, inbound.RegisterInput{
		Email:    "alice2@example.com",
		Username: "alice", // same username
		Password: "password123",
	})

	if !errors.Is(err, ErrUsernameTaken) {
		t.Errorf("expected ErrUsernameTaken, got %v", err)
	}
}

// ── Login ─────────────────────────────────────────────────────────────────────

// helper: registers a user and marks them verified, returns their ID
func registerAndVerify(t *testing.T, svc *authService, userRepo *mockUserRepo, email, username, password string) uuid.UUID {
	t.Helper()
	ctx := context.Background()

	if err := svc.Register(ctx, inbound.RegisterInput{
		Email:    email,
		Username: username,
		Password: password,
	}); err != nil {
		t.Fatalf("register failed: %v", err)
	}

	user, _ := userRepo.GetByEmail(ctx, email)

	// Simulate clicking the verification link
	if err := svc.VerifyEmail(ctx, user.VerifyToken); err != nil {
		t.Fatalf("verify email failed: %v", err)
	}

	return user.ID
}

func TestLogin_Success(t *testing.T) {
	svc, userRepo, _ := newTestAuthService()
	registerAndVerify(t, svc, userRepo, "bob@example.com", "bob", "securepass")

	tokens, err := svc.Login(context.Background(), inbound.LoginInput{
		Email:    "bob@example.com",
		Password: "securepass",
	})

	if err != nil {
		t.Fatalf("expected successful login, got %v", err)
	}
	if tokens.AccessToken == "" || tokens.RefreshToken == "" {
		t.Error("login should return non-empty access and refresh tokens")
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	svc, userRepo, _ := newTestAuthService()
	registerAndVerify(t, svc, userRepo, "bob@example.com", "bob", "securepass")

	_, err := svc.Login(context.Background(), inbound.LoginInput{
		Email:    "bob@example.com",
		Password: "wrongpassword",
	})

	if !errors.Is(err, ErrInvalidCreds) {
		t.Errorf("expected ErrInvalidCreds, got %v", err)
	}
}

func TestLogin_UnverifiedUser(t *testing.T) {
	svc, _, _ := newTestAuthService()
	ctx := context.Background()

	// Register but do NOT verify
	if err := svc.Register(ctx, inbound.RegisterInput{
		Email:    "carol@example.com",
		Username: "carol",
		Password: "pass1234",
	}); err != nil {
		t.Fatalf("register failed: %v", err)
	}

	_, err := svc.Login(ctx, inbound.LoginInput{
		Email:    "carol@example.com",
		Password: "pass1234",
	})

	if !errors.Is(err, ErrNotVerified) {
		t.Errorf("expected ErrNotVerified, got %v", err)
	}
}

func TestLogin_NonExistentEmail(t *testing.T) {
	svc, _, _ := newTestAuthService()

	_, err := svc.Login(context.Background(), inbound.LoginInput{
		Email:    "nobody@example.com",
		Password: "any",
	})

	if !errors.Is(err, ErrInvalidCreds) {
		t.Errorf("expected ErrInvalidCreds for unknown email, got %v", err)
	}
}

// ── VerifyEmail ───────────────────────────────────────────────────────────────

func TestVerifyEmail_Success(t *testing.T) {
	svc, userRepo, _ := newTestAuthService()
	ctx := context.Background()

	if err := svc.Register(ctx, inbound.RegisterInput{
		Email:    "dave@example.com",
		Username: "dave",
		Password: "pass1234",
	}); err != nil {
		t.Fatal(err)
	}

	user, _ := userRepo.GetByEmail(ctx, "dave@example.com")
	token := user.VerifyToken

	if err := svc.VerifyEmail(ctx, token); err != nil {
		t.Fatalf("VerifyEmail failed: %v", err)
	}

	// Reload from repo and check
	updated, _ := userRepo.GetByID(ctx, user.ID)
	if !updated.IsVerified {
		t.Error("user should be verified after VerifyEmail")
	}
	if updated.VerifyToken != "" {
		t.Error("verify token should be cleared after use")
	}
}

func TestVerifyEmail_InvalidToken(t *testing.T) {
	svc, _, _ := newTestAuthService()

	err := svc.VerifyEmail(context.Background(), "totally-fake-token")

	if !errors.Is(err, ErrInvalidToken) {
		t.Errorf("expected ErrInvalidToken, got %v", err)
	}
}

// ── ForgotPassword ────────────────────────────────────────────────────────────

func TestForgotPassword_UnknownEmail_NoError(t *testing.T) {
	// Security: ForgotPassword must NOT return an error for unknown emails.
	// If it did, attackers could enumerate which emails are registered.
	svc, _, _ := newTestAuthService()

	err := svc.ForgotPassword(context.Background(), "unknown@example.com")
	if err != nil {
		t.Errorf("ForgotPassword should return nil for unknown emails to prevent email enumeration, got %v", err)
	}
}

func TestForgotPassword_SendsResetEmail(t *testing.T) {
	svc, userRepo, emailSender := newTestAuthService()
	ctx := context.Background()
	registerAndVerify(t, svc, userRepo, "eve@example.com", "eve", "pass1234")

	if err := svc.ForgotPassword(ctx, "eve@example.com"); err != nil {
		t.Fatalf("ForgotPassword failed: %v", err)
	}

	if len(emailSender.resetsSent) != 1 || emailSender.resetsSent[0] != "eve@example.com" {
		t.Error("password reset email should be sent")
	}
}

// ── ResetPassword ─────────────────────────────────────────────────────────────

func TestResetPassword_Success(t *testing.T) {
	svc, userRepo, _ := newTestAuthService()
	ctx := context.Background()
	registerAndVerify(t, svc, userRepo, "frank@example.com", "frank", "oldpass")

	if err := svc.ForgotPassword(ctx, "frank@example.com"); err != nil {
		t.Fatal(err)
	}

	user, _ := userRepo.GetByEmail(ctx, "frank@example.com")
	resetToken := user.ResetToken

	if err := svc.ResetPassword(ctx, resetToken, "newpass123"); err != nil {
		t.Fatalf("ResetPassword failed: %v", err)
	}

	// Old password should no longer work
	_, err := svc.Login(ctx, inbound.LoginInput{Email: "frank@example.com", Password: "oldpass"})
	if !errors.Is(err, ErrInvalidCreds) {
		t.Error("old password should be rejected after reset")
	}

	// New password should work
	_, err = svc.Login(ctx, inbound.LoginInput{Email: "frank@example.com", Password: "newpass123"})
	if err != nil {
		t.Errorf("new password should work after reset, got %v", err)
	}
}

func TestResetPassword_ExpiredToken(t *testing.T) {
	svc, userRepo, _ := newTestAuthService()
	ctx := context.Background()
	registerAndVerify(t, svc, userRepo, "grace@example.com", "grace", "pass1234")

	if err := svc.ForgotPassword(ctx, "grace@example.com"); err != nil {
		t.Fatal(err)
	}

	// Manually expire the reset token
	user, _ := userRepo.GetByEmail(ctx, "grace@example.com")
	expired := time.Now().Add(-2 * time.Hour) // 2 hours in the past
	user.ResetExpiresAt = &expired
	userRepo.Update(ctx, user)

	err := svc.ResetPassword(ctx, user.ResetToken, "newpass")
	if !errors.Is(err, ErrInvalidToken) {
		t.Errorf("expected ErrInvalidToken for expired token, got %v", err)
	}
}

func TestResetPassword_InvalidToken(t *testing.T) {
	svc, _, _ := newTestAuthService()

	err := svc.ResetPassword(context.Background(), "fake-token", "newpass")
	if !errors.Is(err, ErrInvalidToken) {
		t.Errorf("expected ErrInvalidToken, got %v", err)
	}
}
