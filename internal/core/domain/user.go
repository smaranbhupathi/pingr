package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID             uuid.UUID
	Email          string
	Username       string
	PasswordHash   string
	IsVerified     bool
	VerifyToken    string
	ResetToken     string
	ResetExpiresAt *time.Time
	PlanID         uuid.UUID
	AvatarURL      *string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type Plan struct {
	ID                 uuid.UUID
	Name               string // "free", "pro", etc.
	MaxMonitors        int
	MinIntervalSeconds int // minimum check interval allowed
	CreatedAt          time.Time
}
