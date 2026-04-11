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

var ErrIncidentNotFound = errors.New("incident not found")

type userService struct {
	users         outbound.UserRepository
	plans         outbound.PlanRepository
	alertChannels outbound.AlertChannelRepository
	alertSubs     outbound.AlertSubscriptionRepository
	monitors      outbound.MonitorRepository
	incidents     outbound.IncidentRepository
	email         outbound.EmailSender
	storage       outbound.StorageService // nil when storage is not configured
	notifiers     map[domain.AlertChannelType]outbound.Notifier
}

func NewUserService(
	users outbound.UserRepository,
	plans outbound.PlanRepository,
	alertChannels outbound.AlertChannelRepository,
	alertSubs outbound.AlertSubscriptionRepository,
	monitors outbound.MonitorRepository,
	incidents outbound.IncidentRepository,
	email outbound.EmailSender,
	storage outbound.StorageService,
	notifiers []outbound.Notifier,
) inbound.UserService {
	notifierMap := make(map[domain.AlertChannelType]outbound.Notifier, len(notifiers))
	for _, n := range notifiers {
		notifierMap[n.Type()] = n
	}
	return &userService{
		users:         users,
		plans:         plans,
		alertChannels: alertChannels,
		alertSubs:     alertSubs,
		monitors:      monitors,
		incidents:     incidents,
		email:         email,
		storage:       storage,
		notifiers:     notifierMap,
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
		IsEnabled: true,
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

func (s *userService) ToggleAlertChannel(ctx context.Context, channelID, userID uuid.UUID, enabled bool) error {
	// Verify ownership
	if _, err := s.alertChannels.GetByID(ctx, channelID, userID); err != nil {
		return ErrAlertChannelNotFound
	}
	return s.alertChannels.UpdateEnabled(ctx, channelID, userID, enabled)
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

	// Find the subscribed channel for confirmation message
	var subscribedChannel *domain.AlertChannel
	for i := range channels {
		if channels[i].ID == channelID {
			subscribedChannel = &channels[i]
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

	// Send confirmation — best-effort, never fail the subscription itself
	if subscribedChannel != nil {
		switch subscribedChannel.Type {
		case domain.AlertChannelEmail:
			if emailAddr, ok := subscribedChannel.Config["email"].(string); ok && emailAddr != "" {
				_ = s.email.SendSubscriptionConfirmation(ctx, emailAddr, monitor.Name, monitor.URL)
			}
		default:
			if n, ok := s.notifiers[subscribedChannel.Type]; ok {
				_ = n.SendSubscriptionConfirmation(ctx, monitor.Name, monitor.URL, subscribedChannel.Config)
			}
		}
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

func (s *userService) CreateIncident(ctx context.Context, input inbound.CreateIncidentInput) (*domain.Incident, error) {
	now := time.Now()
	inc := &domain.Incident{
		ID:         uuid.New(),
		UserID:     input.UserID,
		Name:       input.Name,
		Status:     input.Status,
		Source:     "manual",
		MonitorIDs: input.MonitorIDs,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if err := s.incidents.Create(ctx, inc); err != nil {
		return nil, fmt.Errorf("create incident: %w", err)
	}

	update := &domain.IncidentUpdate{
		ID:         uuid.New(),
		IncidentID: inc.ID,
		Status:     input.Status,
		Message:    input.Message,
		Notify:     input.Notify,
		Source:     "manual",
		CreatedAt:  now,
	}
	if err := s.incidents.AddUpdate(ctx, update); err != nil {
		return nil, fmt.Errorf("add initial incident update: %w", err)
	}
	inc.Updates = []domain.IncidentUpdate{*update}

	return inc, nil
}

func (s *userService) GetIncident(ctx context.Context, id, userID uuid.UUID) (*domain.Incident, error) {
	inc, err := s.incidents.GetByID(ctx, id, userID)
	if err != nil {
		return nil, ErrIncidentNotFound
	}
	return inc, nil
}

func (s *userService) ListIncidents(ctx context.Context, userID uuid.UUID) ([]domain.Incident, error) {
	return s.incidents.ListByUser(ctx, userID)
}

func (s *userService) PostIncidentUpdate(ctx context.Context, input inbound.PostIncidentUpdateInput) (*domain.Incident, error) {
	inc, err := s.incidents.GetByID(ctx, input.IncidentID, input.UserID)
	if err != nil {
		return nil, ErrIncidentNotFound
	}

	now := time.Now()
	update := &domain.IncidentUpdate{
		ID:         uuid.New(),
		IncidentID: inc.ID,
		Status:     input.Status,
		Message:    input.Message,
		Notify:     input.Notify,
		Source:     "manual",
		CreatedAt:  now,
	}
	if err := s.incidents.AddUpdate(ctx, update); err != nil {
		return nil, fmt.Errorf("add incident update: %w", err)
	}

	var resolvedAt *time.Time
	if input.Status == domain.IncidentStatusResolved {
		resolvedAt = &now
	}
	if err := s.incidents.UpdateStatus(ctx, inc.ID, input.Status, resolvedAt); err != nil {
		return nil, fmt.Errorf("update incident status: %w", err)
	}

	return s.incidents.GetByID(ctx, inc.ID, input.UserID)
}
