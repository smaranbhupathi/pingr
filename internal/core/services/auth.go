package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/smaranbhupathi/pingr/internal/core/domain"
	"github.com/smaranbhupathi/pingr/internal/core/ports/inbound"
	"github.com/smaranbhupathi/pingr/internal/core/ports/outbound"
)

var (
	ErrEmailTaken    = errors.New("email already registered")
	ErrUsernameTaken = errors.New("username already taken")
	ErrInvalidCreds  = errors.New("invalid email or password")
	ErrNotVerified   = errors.New("email not verified")
	ErrInvalidToken  = errors.New("invalid or expired token")
)

type jwtClaims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

type AuthServiceConfig struct {
	JWTSecret            string
	AccessTokenDuration  time.Duration
	RefreshTokenDuration time.Duration
	AppBaseURL           string
}

type authService struct {
	users         outbound.UserRepository
	plans         outbound.PlanRepository
	alertChannels outbound.AlertChannelRepository
	email         outbound.EmailSender
	config        AuthServiceConfig
}

func NewAuthService(
	users outbound.UserRepository,
	plans outbound.PlanRepository,
	alertChannels outbound.AlertChannelRepository,
	email outbound.EmailSender,
	config AuthServiceConfig,
) inbound.AuthService {
	return &authService{users: users, plans: plans, alertChannels: alertChannels, email: email, config: config}
}

func (s *authService) Register(ctx context.Context, input inbound.RegisterInput) error {
	if _, err := s.users.GetByEmail(ctx, input.Email); err == nil {
		return ErrEmailTaken
	}
	if _, err := s.users.GetByUsername(ctx, input.Username); err == nil {
		return ErrUsernameTaken
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	freePlan, err := s.plans.GetByName(ctx, "free")
	if err != nil {
		return fmt.Errorf("get free plan: %w", err)
	}

	verifyToken, err := generateToken()
	if err != nil {
		return err
	}

	user := &domain.User{
		ID:           uuid.New(),
		Email:        input.Email,
		Username:     input.Username,
		PasswordHash: string(hash),
		IsVerified:   false,
		VerifyToken:  verifyToken,
		PlanID:       freePlan.ID,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.users.Create(ctx, user); err != nil {
		// Postgres unique constraint violation (SQLSTATE 23505).
		// This is a fallback for the rare race where two requests slip past
		// the GetByEmail/GetByUsername checks above simultaneously.
		if strings.Contains(err.Error(), "23505") {
			if strings.Contains(err.Error(), "email") {
				return ErrEmailTaken
			}
			return ErrUsernameTaken
		}
		return fmt.Errorf("create user: %w", err)
	}

	// Auto-create default email alert channel so new users get alerts immediately
	defaultChannel := &domain.AlertChannel{
		ID:        uuid.New(),
		UserID:    user.ID,
		Type:      domain.AlertChannelEmail,
		Config:    map[string]any{"email": user.Email},
		IsDefault: true,
		CreatedAt: time.Now(),
	}
	if err := s.alertChannels.Create(ctx, defaultChannel); err != nil {
		return fmt.Errorf("create default alert channel: %w", err)
	}

	return s.email.SendVerification(ctx, user.Email, verifyToken)
}

func (s *authService) Login(ctx context.Context, input inbound.LoginInput) (*inbound.AuthTokens, error) {
	user, err := s.users.GetByEmail(ctx, input.Email)
	if err != nil {
		return nil, ErrInvalidCreds
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return nil, ErrInvalidCreds
	}

	if !user.IsVerified {
		return nil, ErrNotVerified
	}

	return s.issueTokens(user.ID)
}

func (s *authService) RefreshTokens(ctx context.Context, refreshToken string) (*inbound.AuthTokens, error) {
	claims, err := s.parseToken(refreshToken)
	if err != nil {
		return nil, ErrInvalidToken
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, ErrInvalidToken
	}

	if _, err := s.users.GetByID(ctx, userID); err != nil {
		return nil, ErrInvalidToken
	}

	return s.issueTokens(userID)
}

func (s *authService) VerifyEmail(ctx context.Context, token string) error {
	user, err := s.users.GetByVerifyToken(ctx, token)
	if err != nil {
		return ErrInvalidToken
	}

	user.IsVerified = true
	user.VerifyToken = ""
	user.UpdatedAt = time.Now()

	return s.users.Update(ctx, user)
}

func (s *authService) ForgotPassword(ctx context.Context, email string) error {
	user, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		// Don't reveal whether email exists
		return nil
	}

	token, err := generateToken()
	if err != nil {
		return err
	}

	expires := time.Now().Add(1 * time.Hour)
	user.ResetToken = token
	user.ResetExpiresAt = &expires
	user.UpdatedAt = time.Now()

	if err := s.users.Update(ctx, user); err != nil {
		return err
	}

	return s.email.SendPasswordReset(ctx, user.Email, token)
}

func (s *authService) ResetPassword(ctx context.Context, token, newPassword string) error {
	user, err := s.users.GetByResetToken(ctx, token)
	if err != nil {
		return ErrInvalidToken
	}

	if user.ResetExpiresAt == nil || time.Now().After(*user.ResetExpiresAt) {
		return ErrInvalidToken
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.PasswordHash = string(hash)
	user.ResetToken = ""
	user.ResetExpiresAt = nil
	user.UpdatedAt = time.Now()

	return s.users.Update(ctx, user)
}

func (s *authService) issueTokens(userID uuid.UUID) (*inbound.AuthTokens, error) {
	access, err := s.signToken(userID, s.config.AccessTokenDuration)
	if err != nil {
		return nil, err
	}

	refresh, err := s.signToken(userID, s.config.RefreshTokenDuration)
	if err != nil {
		return nil, err
	}

	return &inbound.AuthTokens{AccessToken: access, RefreshToken: refresh}, nil
}

func (s *authService) signToken(userID uuid.UUID, duration time.Duration) (string, error) {
	claims := jwtClaims{
		UserID: userID.String(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(s.config.JWTSecret))
}

func (s *authService) parseToken(tokenStr string) (*jwtClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &jwtClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(s.config.JWTSecret), nil
	})
	if err != nil || !token.Valid {
		return nil, ErrInvalidToken
	}
	claims, ok := token.Claims.(*jwtClaims)
	if !ok {
		return nil, ErrInvalidToken
	}
	return claims, nil
}

func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}
	return hex.EncodeToString(b), nil
}
