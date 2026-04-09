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
	ErrAlertChannelNotFound = errors.New("alert channel not found")
	ErrAlreadySubscribed    = errors.New("monitor already subscribed to this channel")
)

type userService struct {
	users         outbound.UserRepository
	plans         outbound.PlanRepository
	alertChannels outbound.AlertChannelRepository
	alertSubs     outbound.AlertSubscriptionRepository
	monitors      outbound.MonitorRepository
}

func NewUserService(
	users outbound.UserRepository,
	plans outbound.PlanRepository,
	alertChannels outbound.AlertChannelRepository,
	alertSubs outbound.AlertSubscriptionRepository,
	monitors outbound.MonitorRepository,
) inbound.UserService {
	return &userService{
		users:         users,
		plans:         plans,
		alertChannels: alertChannels,
		alertSubs:     alertSubs,
		monitors:      monitors,
	}
}

func (s *userService) GetProfile(ctx context.Context, userID uuid.UUID) (*inbound.UserProfile, error) {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}

	plan, err := s.plans.GetByID(ctx, user.PlanID)
	if err != nil {
		return nil, fmt.Errorf("get plan: %w", err)
	}

	return &inbound.UserProfile{
		ID:        user.ID,
		Email:     user.Email,
		Username:  user.Username,
		Plan:      plan.Name,
		CreatedAt: user.CreatedAt,
	}, nil
}

func (s *userService) CreateAlertChannel(ctx context.Context, input inbound.CreateAlertChannelInput) (*domain.AlertChannel, error) {
	ch := &domain.AlertChannel{
		ID:        uuid.New(),
		UserID:    input.UserID,
		Type:      input.Type,
		Config:    input.Config,
		IsDefault: input.IsDefault,
		CreatedAt: time.Now(),
	}
	if err := s.alertChannels.Create(ctx, ch); err != nil {
		return nil, fmt.Errorf("create alert channel: %w", err)
	}
	return ch, nil
}

func (s *userService) ListAlertChannels(ctx context.Context, userID uuid.UUID) ([]domain.AlertChannel, error) {
	return s.alertChannels.GetByUserID(ctx, userID)
}

func (s *userService) DeleteAlertChannel(ctx context.Context, channelID, userID uuid.UUID) error {
	channels, err := s.alertChannels.GetByUserID(ctx, userID)
	if err != nil {
		return err
	}
	for _, ch := range channels {
		if ch.ID == channelID {
			return s.alertChannels.Delete(ctx, channelID)
		}
	}
	return ErrAlertChannelNotFound
}

func (s *userService) SubscribeMonitorToChannel(ctx context.Context, monitorID, channelID, userID uuid.UUID) error {
	// Verify monitor belongs to user
	monitor, err := s.monitors.GetByID(ctx, monitorID)
	if err != nil || monitor.UserID != userID {
		return ErrMonitorNotFound
	}

	// Verify channel belongs to user
	channels, err := s.alertChannels.GetByUserID(ctx, userID)
	if err != nil {
		return err
	}
	found := false
	for _, ch := range channels {
		if ch.ID == channelID {
			found = true
			break
		}
	}
	if !found {
		return ErrAlertChannelNotFound
	}

	sub := &domain.AlertSubscription{
		ID:             uuid.New(),
		MonitorID:      monitorID,
		AlertChannelID: channelID,
		CreatedAt:      time.Now(),
	}
	return s.alertSubs.Create(ctx, sub)
}
