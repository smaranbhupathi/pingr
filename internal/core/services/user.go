package services

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
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

var ErrComponentNotFound = errors.New("component not found")

type userService struct {
	users         outbound.UserRepository
	plans         outbound.PlanRepository
	alertChannels outbound.AlertChannelRepository
	alertSubs     outbound.AlertSubscriptionRepository
	monitors      outbound.MonitorRepository
	components    outbound.ComponentRepository
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
	components outbound.ComponentRepository,
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
		components:    components,
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

	// Apply component_status overrides if provided.
	s.applyMonitorStatuses(ctx, input.MonitorStatuses)

	if input.Notify {
		// Re-fetch so Monitors (name/URL) are populated for the notification body.
		full, err := s.incidents.GetByID(ctx, inc.ID, input.UserID)
		if err != nil {
			slog.Error("fanout: failed to re-fetch incident for notification", "incident_id", inc.ID, "error", err)
		} else {
			s.fanoutIncidentUpdate(ctx, *full, *update)
		}
	}

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

	// Apply component_status overrides if provided.
	s.applyMonitorStatuses(ctx, input.MonitorStatuses)

	// Fetch the full updated incident (with monitors populated) before notifying.
	updated, err := s.incidents.GetByID(ctx, inc.ID, input.UserID)
	if err != nil {
		return nil, err
	}

	if input.Notify {
		s.fanoutIncidentUpdate(ctx, *updated, *update)
	}

	return updated, nil
}

// fanoutIncidentUpdate sends the incident update to every alert channel subscribed
// to any monitor affected by this incident. Channels are deduplicated — if the same
// Slack webhook is subscribed to two affected monitors, it only gets one message.
func (s *userService) fanoutIncidentUpdate(ctx context.Context, incident domain.Incident, update domain.IncidentUpdate) {
	if len(incident.MonitorIDs) == 0 {
		slog.Warn("fanout: incident has no affected monitors, skipping notification", "incident_id", incident.ID)
		return
	}

	seen := make(map[uuid.UUID]bool)

	for _, monitorID := range incident.MonitorIDs {
		channels, err := s.alertChannels.GetByMonitorID(ctx, monitorID)
		if err != nil {
			slog.Error("fanout: get channels failed", "incident_id", incident.ID, "monitor_id", monitorID, "error", err)
			continue
		}
		if len(channels) == 0 {
			slog.Warn("fanout: no subscribed channels for monitor", "incident_id", incident.ID, "monitor_id", monitorID)
			continue
		}
		for _, ch := range channels {
			if seen[ch.ID] {
				continue
			}
			if !ch.IsEnabled {
				slog.Debug("fanout: channel disabled, skipping", "channel_id", ch.ID, "type", ch.Type)
				continue
			}
			seen[ch.ID] = true

			notifier, ok := s.notifiers[ch.Type]
			if !ok {
				slog.Warn("fanout: no notifier registered for channel type", "type", ch.Type, "channel_id", ch.ID)
				continue
			}
			if err := notifier.SendIncidentUpdate(ctx, incident, update, ch.Config); err != nil {
				slog.Error("fanout: send failed", "channel_id", ch.ID, "type", ch.Type, "incident_id", incident.ID, "error", err)
			} else {
				slog.Info("fanout: notification sent", "channel_id", ch.ID, "type", ch.Type, "incident_id", incident.ID)
			}
		}
	}
}

// applyMonitorStatuses updates the component_status of affected monitors.
func (s *userService) applyMonitorStatuses(ctx context.Context, statuses map[uuid.UUID]domain.ComponentStatus) {
	for monitorID, status := range statuses {
		if err := s.monitors.UpdateComponentStatus(ctx, monitorID, status); err != nil {
			slog.Error("apply monitor status: update failed", "monitor_id", monitorID, "status", status, "error", err)
		}
	}
}

// ── Components ────────────────────────────────────────────────────────────────

func (s *userService) CreateComponent(ctx context.Context, input inbound.CreateComponentInput) (*domain.Component, error) {
	now := time.Now()
	c := &domain.Component{
		ID:          uuid.New(),
		UserID:      input.UserID,
		Name:        input.Name,
		Description: input.Description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.components.Create(ctx, c); err != nil {
		return nil, fmt.Errorf("create component: %w", err)
	}
	return c, nil
}

func (s *userService) ListComponents(ctx context.Context, userID uuid.UUID) ([]domain.Component, error) {
	return s.components.ListByUser(ctx, userID)
}

func (s *userService) UpdateComponent(ctx context.Context, id, userID uuid.UUID, input inbound.UpdateComponentInput) (*domain.Component, error) {
	c, err := s.components.GetByID(ctx, id, userID)
	if err != nil {
		return nil, ErrComponentNotFound
	}
	if input.Name != nil {
		c.Name = *input.Name
	}
	if input.Description != nil {
		c.Description = *input.Description
	}
	if err := s.components.Update(ctx, c); err != nil {
		return nil, fmt.Errorf("update component: %w", err)
	}
	return c, nil
}

func (s *userService) DeleteComponent(ctx context.Context, id, userID uuid.UUID) error {
	if _, err := s.components.GetByID(ctx, id, userID); err != nil {
		return ErrComponentNotFound
	}
	return s.components.Delete(ctx, id, userID)
}

// ── Import / Export ───────────────────────────────────────────────────────────

func (s *userService) ImportAlertChannels(ctx context.Context, userID uuid.UUID, rows []inbound.ImportChannelRow, onConflict string) (*inbound.ImportResult, error) {
	existing, err := s.alertChannels.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("load existing channels: %w", err)
	}

	// Build conflict lookup: "type:value" → existing channel
	lookup := make(map[string]*domain.AlertChannel, len(existing))
	for i := range existing {
		ch := &existing[i]
		lookup[importConflictKey(ch.Type, ch.Config)] = ch
	}

	result := &inbound.ImportResult{Errors: []inbound.ImportError{}}

	for i, row := range rows {
		rowNum := i + 1

		if err := validateImportRow(row); err != nil {
			result.Errors = append(result.Errors, inbound.ImportError{Row: rowNum, Name: row.Name, Reason: err.Error()})
			continue
		}

		key := string(row.Type) + ":" + row.Value
		conflicting, hasConflict := lookup[key]

		if hasConflict {
			if onConflict == "skip" {
				result.Skipped++
				continue
			}
			// overwrite: update name + enabled on the existing channel
			if err := s.alertChannels.UpdateName(ctx, conflicting.ID, userID, row.Name); err != nil {
				result.Errors = append(result.Errors, inbound.ImportError{Row: rowNum, Name: row.Name, Reason: "failed to update name"})
				continue
			}
			if err := s.alertChannels.UpdateEnabled(ctx, conflicting.ID, userID, row.Enabled); err != nil {
				result.Errors = append(result.Errors, inbound.ImportError{Row: rowNum, Name: row.Name, Reason: "failed to update enabled"})
				continue
			}
			result.Overwritten++
			continue
		}

		config := make(map[string]any)
		if row.Type == domain.AlertChannelEmail {
			config["email"] = row.Value
		} else {
			config["webhook_url"] = row.Value
		}

		ch := &domain.AlertChannel{
			ID:        uuid.New(),
			UserID:    userID,
			Name:      row.Name,
			Type:      row.Type,
			Config:    config,
			IsDefault: false,
			IsEnabled: row.Enabled,
			CreatedAt: time.Now(),
		}
		if err := s.alertChannels.Create(ctx, ch); err != nil {
			result.Errors = append(result.Errors, inbound.ImportError{Row: rowNum, Name: row.Name, Reason: "failed to save"})
			continue
		}
		result.Imported++
	}

	return result, nil
}

func importConflictKey(t domain.AlertChannelType, config map[string]any) string {
	var val string
	if t == domain.AlertChannelEmail {
		val, _ = config["email"].(string)
	} else {
		val, _ = config["webhook_url"].(string)
	}
	return string(t) + ":" + val
}

func validateImportRow(row inbound.ImportChannelRow) error {
	if row.Name == "" {
		return fmt.Errorf("name is required")
	}
	switch row.Type {
	case domain.AlertChannelEmail:
		if row.Value == "" || !strings.Contains(row.Value, "@") {
			return fmt.Errorf("invalid email address")
		}
	case domain.AlertChannelSlack, domain.AlertChannelDiscord:
		if row.Value == "" || !strings.HasPrefix(row.Value, "https://") {
			return fmt.Errorf("invalid webhook URL")
		}
	default:
		return fmt.Errorf("unsupported type %q — must be email, slack, or discord", row.Type)
	}
	return nil
}

// ── Monitor meta ──────────────────────────────────────────────────────────────

func (s *userService) UpdateMonitorMeta(ctx context.Context, id, userID uuid.UUID, name, description string, componentID *uuid.UUID) (*domain.Monitor, error) {
	monitor, err := s.monitors.GetByID(ctx, id)
	if err != nil || monitor.UserID != userID {
		return nil, ErrMonitorNotFound
	}
	if name != "" {
		monitor.Name = name
	}
	monitor.Description = description
	monitor.ComponentID = componentID
	monitor.UpdatedAt = time.Now()
	if err := s.monitors.Update(ctx, monitor); err != nil {
		return nil, fmt.Errorf("update monitor meta: %w", err)
	}
	return monitor, nil
}
