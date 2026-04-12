package inbound

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/smaranbhupathi/pingr/internal/core/domain"
)

type UserProfile struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	Username  string    `json:"username"`
	Plan      string    `json:"plan"`
	AvatarURL *string   `json:"avatar_url"`
	CreatedAt time.Time `json:"created_at"`
}

// AvatarUploadResult is returned by AvatarUploadURL.
// upload_url: the browser PUTs the file here (goes straight to R2/S3).
// public_url: the permanent URL to store after the upload succeeds.
type AvatarUploadResult struct {
	UploadURL string `json:"upload_url"`
	PublicURL string `json:"public_url"`
}

type CreateAlertChannelInput struct {
	UserID    uuid.UUID
	Name      string
	Type      domain.AlertChannelType
	Config    map[string]any
	IsDefault bool
}

func (i CreateAlertChannelInput) Validate() map[string]string {
	errs := map[string]string{}
	if i.Name == "" {
		errs["name"] = "required"
	}
	switch i.Type {
	case domain.AlertChannelEmail:
		email, ok := i.Config["email"].(string)
		if !ok || email == "" {
			errs["config.email"] = "required"
		} else if len(email) < 3 {
			errs["config.email"] = "invalid email"
		}
	case domain.AlertChannelSlack, domain.AlertChannelDiscord:
		url, ok := i.Config["webhook_url"].(string)
		if !ok || url == "" {
			errs["config.webhook_url"] = "required"
		} else if len(url) < 10 {
			errs["config.webhook_url"] = "invalid webhook URL"
		}
	default:
		errs["type"] = "unsupported channel type"
	}
	return errs
}

type CreateIncidentInput struct {
	UserID           uuid.UUID
	Name             string
	Status           domain.IncidentStatus
	Message          string
	MonitorIDs       []uuid.UUID
	MonitorStatuses  map[uuid.UUID]domain.ComponentStatus // optional: set component_status per monitor
	Notify           bool
}

type PostIncidentUpdateInput struct {
	IncidentID      uuid.UUID
	UserID          uuid.UUID
	Status          domain.IncidentStatus
	Message         string
	MonitorStatuses map[uuid.UUID]domain.ComponentStatus // optional: set component_status per monitor
	Notify          bool
}

type CreateComponentInput struct {
	UserID      uuid.UUID
	Name        string
	Description string
}

type UpdateComponentInput struct {
	Name        *string
	Description *string
}

type UserService interface {
	GetProfile(ctx context.Context, userID uuid.UUID) (*UserProfile, error)
	// AvatarUploadURL generates a short-lived presigned PUT URL for the browser
	// to upload an avatar directly to object storage (S3/R2/MinIO).
	AvatarUploadURL(ctx context.Context, userID uuid.UUID, contentType string) (*AvatarUploadResult, error)
	// UpdateAvatar persists the public URL returned after a successful upload.
	UpdateAvatar(ctx context.Context, userID uuid.UUID, publicURL string) error
	CreateAlertChannel(ctx context.Context, input CreateAlertChannelInput) (*domain.AlertChannel, error)
	GetAlertChannel(ctx context.Context, channelID, userID uuid.UUID) (*domain.AlertChannel, error)
	ListAlertChannels(ctx context.Context, userID uuid.UUID) ([]domain.AlertChannel, error)
	UpdateAlertChannelName(ctx context.Context, channelID, userID uuid.UUID, name string) error
	ToggleAlertChannel(ctx context.Context, channelID, userID uuid.UUID, enabled bool) error
	DeleteAlertChannel(ctx context.Context, channelID, userID uuid.UUID) error
	SubscribeMonitorToChannel(ctx context.Context, monitorID, channelID, userID uuid.UUID) error
	UnsubscribeMonitorFromChannel(ctx context.Context, monitorID, channelID, userID uuid.UUID) error
	ListMonitorSubscriptions(ctx context.Context, monitorID, userID uuid.UUID) ([]domain.AlertChannel, error)

	// Incidents
	CreateIncident(ctx context.Context, input CreateIncidentInput) (*domain.Incident, error)
	GetIncident(ctx context.Context, id, userID uuid.UUID) (*domain.Incident, error)
	ListIncidents(ctx context.Context, userID uuid.UUID) ([]domain.Incident, error)
	PostIncidentUpdate(ctx context.Context, input PostIncidentUpdateInput) (*domain.Incident, error)

	// Components
	CreateComponent(ctx context.Context, input CreateComponentInput) (*domain.Component, error)
	ListComponents(ctx context.Context, userID uuid.UUID) ([]domain.Component, error)
	UpdateComponent(ctx context.Context, id, userID uuid.UUID, input UpdateComponentInput) (*domain.Component, error)
	DeleteComponent(ctx context.Context, id, userID uuid.UUID) error

	// Monitor edits
	UpdateMonitorMeta(ctx context.Context, id, userID uuid.UUID, name, description string, componentID *uuid.UUID) (*domain.Monitor, error)
}
