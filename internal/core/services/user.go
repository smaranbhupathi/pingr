package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
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
	email         outbound.EmailSender
	storage       outbound.StorageService // nil when storage is not configured
}

func NewUserService(
	users outbound.UserRepository,
	plans outbound.PlanRepository,
	alertChannels outbound.AlertChannelRepository,
	alertSubs outbound.AlertSubscriptionRepository,
	monitors outbound.MonitorRepository,
	email outbound.EmailSender,
	storage outbound.StorageService,
) inbound.UserService {
	return &userService{
		users:         users,
		plans:         plans,
		alertChannels: alertChannels,
		alertSubs:     alertSubs,
		monitors:      monitors,
		email:         email,
		storage:       storage,
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
		AvatarURL: user.AvatarURL,
		CreatedAt: user.CreatedAt,
	}, nil
}

var ErrStorageNotConfigured = errors.New("storage not configured")

// allowedAvatarTypes maps accepted content-types to file extensions.
// We validate here in the core so no adapter needs to duplicate this check.
var allowedAvatarTypes = map[string]string{
	"image/jpeg": "jpg",
	"image/png":  "png",
	"image/webp": "webp",
}

func (s *userService) AvatarUploadURL(ctx context.Context, userID uuid.UUID, contentType string) (*inbound.AvatarUploadResult, error) {
	if s.storage == nil {
		return nil, ErrStorageNotConfigured
	}

	ext, ok := allowedAvatarTypes[strings.ToLower(contentType)]
	if !ok {
		return nil, fmt.Errorf("unsupported image type: %s", contentType)
	}

	// Key format: avatars/<userID>.<ext>
	// Using the userID as the filename means a new upload naturally overwrites
	// the old one — no orphaned files accumulate in the bucket.
	key := fmt.Sprintf("avatars/%s.%s", userID, ext)

	uploadURL, err := s.storage.PresignPUT(ctx, key, 5*time.Minute)
	if err != nil {
		return nil, fmt.Errorf("presign: %w", err)
	}

	return &inbound.AvatarUploadResult{
		UploadURL: uploadURL,
		PublicURL: s.storage.PublicURL(key),
	}, nil
}

func (s *userService) UpdateAvatar(ctx context.Context, userID uuid.UUID, publicURL string) error {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("get user: %w", err)
	}
	user.AvatarURL = &publicURL
	user.UpdatedAt = time.Now()
	return s.users.Update(ctx, user)
}

func (s *userService) CreateAlertChannel(ctx context.Context, input inbound.CreateAlertChannelInput) (*domain.AlertChannel, error) {
	ch := &domain.AlertChannel{
		ID:        uuid.New(),
		UserID:    input.UserID,
		Name:      input.Name,
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

func (s *userService) GetAlertChannel(ctx context.Context, channelID, userID uuid.UUID) (*domain.AlertChannel, error) {
	ch, err := s.alertChannels.GetByID(ctx, channelID, userID)
	if err != nil {
		return nil, ErrAlertChannelNotFound
	}
	return ch, nil
}

func (s *userService) ListAlertChannels(ctx context.Context, userID uuid.UUID) ([]domain.AlertChannel, error) {
	return s.alertChannels.GetByUserID(ctx, userID)
}

func (s *userService) UpdateAlertChannelName(ctx context.Context, channelID, userID uuid.UUID, name string) error {
	if name == "" {
		return fmt.Errorf("name is required")
	}
	if err := s.alertChannels.UpdateName(ctx, channelID, userID, name); err != nil {
		return fmt.Errorf("update alert channel name: %w", err)
	}
	return nil
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

func (s *userService) ListMonitorSubscriptions(ctx context.Context, monitorID, userID uuid.UUID) ([]domain.AlertChannel, error) {
	monitor, err := s.monitors.GetByID(ctx, monitorID)
	if err != nil || monitor.UserID != userID {
		return nil, ErrMonitorNotFound
	}
	return s.alertChannels.GetByMonitorID(ctx, monitorID)
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

	// Find the channel to get the email address for the confirmation
	var channelEmail string
	for _, ch := range channels {
		if ch.ID == channelID {
			if e, ok := ch.Config["email"].(string); ok {
				channelEmail = e
			}
			break
		}
	}

	sub := &domain.AlertSubscription{
		ID:             uuid.New(),
		MonitorID:      monitorID,
		AlertChannelID: channelID,
		CreatedAt:      time.Now(),
	}
	if err := s.alertSubs.Create(ctx, sub); err != nil {
		return err
	}

	// Send confirmation email (best-effort — don't fail the subscription if email fails)
	if channelEmail != "" {
		_ = s.email.SendSubscriptionConfirmation(ctx, channelEmail, monitor.Name, monitor.URL)
	}
	return nil
}

func (s *userService) UnsubscribeMonitorFromChannel(ctx context.Context, monitorID, channelID, userID uuid.UUID) error {
	monitor, err := s.monitors.GetByID(ctx, monitorID)
	if err != nil || monitor.UserID != userID {
		return ErrMonitorNotFound
	}

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

	return s.alertSubs.DeleteByMonitorAndChannel(ctx, monitorID, channelID)
}
