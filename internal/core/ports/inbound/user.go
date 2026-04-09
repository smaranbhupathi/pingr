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
	CreatedAt time.Time `json:"created_at"`
}

type CreateAlertChannelInput struct {
	UserID    uuid.UUID
	Type      domain.AlertChannelType
	Config    map[string]any
	IsDefault bool
}

func (i CreateAlertChannelInput) Validate() map[string]string {
	errs := map[string]string{}
	switch i.Type {
	case domain.AlertChannelEmail:
		email, ok := i.Config["email"].(string)
		if !ok || email == "" {
			errs["config.email"] = "required"
		} else if len(email) < 3 {
			errs["config.email"] = "invalid email"
		}
	default:
		errs["type"] = "unsupported channel type"
	}
	return errs
}

type UserService interface {
	GetProfile(ctx context.Context, userID uuid.UUID) (*UserProfile, error)
	CreateAlertChannel(ctx context.Context, input CreateAlertChannelInput) (*domain.AlertChannel, error)
	ListAlertChannels(ctx context.Context, userID uuid.UUID) ([]domain.AlertChannel, error)
	DeleteAlertChannel(ctx context.Context, channelID, userID uuid.UUID) error
	SubscribeMonitorToChannel(ctx context.Context, monitorID, channelID, userID uuid.UUID) error
}
