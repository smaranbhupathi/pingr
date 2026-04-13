package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/smaranbhupathi/pingr/internal/adapters/inbound/http/middleware"
	"github.com/smaranbhupathi/pingr/internal/config"
	"github.com/smaranbhupathi/pingr/internal/core/domain"
	"github.com/smaranbhupathi/pingr/internal/core/ports/inbound"
	"github.com/smaranbhupathi/pingr/internal/core/services"
)

type UserHandler struct {
	users inbound.UserService
	cfg   *config.Config
	log   *slog.Logger
}

func NewUserHandler(users inbound.UserService, cfg *config.Config, log *slog.Logger) *UserHandler {
	return &UserHandler{users: users, cfg: cfg, log: log}
}

func (h *UserHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	profile, err := h.users.GetProfile(r.Context(), userID)
	if err != nil {
		h.log.ErrorContext(r.Context(), "get profile failed",
			"request_id", middleware.RequestIDFromContext(r.Context()),
			"user_id", userID,
			"error", err,
		)
		Error(w, http.StatusInternalServerError, "failed to get profile")
		return
	}

	JSON(w, http.StatusOK, profile)
}

func (h *UserHandler) CreateAlertChannel(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var body struct {
		Name      string                  `json:"name"`
		Type      domain.AlertChannelType `json:"type"`
		Config    map[string]any          `json:"config"`
		IsDefault bool                    `json:"is_default"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Feature-flag check — refuse creation if the channel type is disabled in config.yaml
	switch body.Type {
	case domain.AlertChannelEmail:
		if !h.cfg.Features.EmailAlerts {
			Error(w, http.StatusForbidden, "email alerts are currently disabled")
			return
		}
	case domain.AlertChannelSlack:
		if !h.cfg.Features.SlackAlerts {
			Error(w, http.StatusForbidden, "slack alerts are currently disabled")
			return
		}
	case domain.AlertChannelDiscord:
		if !h.cfg.Features.DiscordAlerts {
			Error(w, http.StatusForbidden, "discord alerts are currently disabled")
			return
		}
	}

	input := inbound.CreateAlertChannelInput{
		UserID:    userID,
		Name:      body.Name,
		Type:      body.Type,
		Config:    body.Config,
		IsDefault: body.IsDefault,
	}
	if errs := input.Validate(); len(errs) > 0 {
		JSON(w, http.StatusUnprocessableEntity, map[string]any{"errors": errs})
		return
	}

	ch, err := h.users.CreateAlertChannel(r.Context(), input)
	if err != nil {
		h.log.ErrorContext(r.Context(), "create alert channel failed",
			"request_id", middleware.RequestIDFromContext(r.Context()),
			"user_id", userID,
			"type", body.Type,
			"error", err,
		)
		Error(w, http.StatusInternalServerError, "failed to create alert channel")
		return
	}

	h.log.InfoContext(r.Context(), "alert channel created",
		"request_id", middleware.RequestIDFromContext(r.Context()),
		"user_id", userID,
		"channel_id", ch.ID,
		"type", ch.Type,
	)
	JSON(w, http.StatusCreated, ch)
}

func (h *UserHandler) ListAlertChannels(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	channels, err := h.users.ListAlertChannels(r.Context(), userID)
	if err != nil {
		h.log.ErrorContext(r.Context(), "list alert channels failed",
			"request_id", middleware.RequestIDFromContext(r.Context()),
			"user_id", userID,
			"error", err,
		)
		Error(w, http.StatusInternalServerError, "failed to list alert channels")
		return
	}

	JSON(w, http.StatusOK, channels)
}

func (h *UserHandler) GetAlertChannel(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	channelID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid channel id")
		return
	}

	ch, err := h.users.GetAlertChannel(r.Context(), channelID, userID)
	if err != nil {
		if errors.Is(err, services.ErrAlertChannelNotFound) {
			Error(w, http.StatusNotFound, "alert channel not found")
			return
		}
		Error(w, http.StatusInternalServerError, "failed to get alert channel")
		return
	}

	JSON(w, http.StatusOK, ch)
}

func (h *UserHandler) UpdateAlertChannel(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	channelID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid channel id")
		return
	}

	// Both fields are optional — send only what you want to change.
	var body struct {
		Name      *string `json:"name"`
		IsEnabled *bool   `json:"is_enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if body.Name != nil {
		if err := h.users.UpdateAlertChannelName(r.Context(), channelID, userID, *body.Name); err != nil {
			if errors.Is(err, services.ErrAlertChannelNotFound) {
				Error(w, http.StatusNotFound, "alert channel not found")
				return
			}
			Error(w, http.StatusBadRequest, err.Error())
			return
		}
	}

	if body.IsEnabled != nil {
		if err := h.users.ToggleAlertChannel(r.Context(), channelID, userID, *body.IsEnabled); err != nil {
			if errors.Is(err, services.ErrAlertChannelNotFound) {
				Error(w, http.StatusNotFound, "alert channel not found")
				return
			}
			Error(w, http.StatusInternalServerError, "failed to update channel")
			return
		}
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *UserHandler) DeleteAlertChannel(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	channelID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid channel id")
		return
	}

	if err := h.users.DeleteAlertChannel(r.Context(), channelID, userID); err != nil {
		if errors.Is(err, services.ErrAlertChannelNotFound) {
			Error(w, http.StatusNotFound, "alert channel not found")
			return
		}
		h.log.ErrorContext(r.Context(), "delete alert channel failed",
			"request_id", middleware.RequestIDFromContext(r.Context()),
			"user_id", userID,
			"channel_id", channelID,
			"error", err,
		)
		Error(w, http.StatusInternalServerError, "failed to delete alert channel")
		return
	}

	h.log.InfoContext(r.Context(), "alert channel deleted",
		"request_id", middleware.RequestIDFromContext(r.Context()),
		"user_id", userID,
		"channel_id", channelID,
	)
	w.WriteHeader(http.StatusNoContent)
}

func (h *UserHandler) ListMonitorSubscriptions(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	monitorID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid monitor id")
		return
	}

	channels, err := h.users.ListMonitorSubscriptions(r.Context(), monitorID, userID)
	if err != nil {
		h.log.ErrorContext(r.Context(), "list monitor subscriptions failed",
			"request_id", middleware.RequestIDFromContext(r.Context()),
			"user_id", userID,
			"monitor_id", monitorID,
			"error", err,
		)
		Error(w, http.StatusInternalServerError, "failed to list subscriptions")
		return
	}

	JSON(w, http.StatusOK, channels)
}

func (h *UserHandler) UnsubscribeMonitorFromChannel(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	monitorID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid monitor id")
		return
	}

	channelID, err := uuid.Parse(chi.URLParam(r, "channelId"))
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid channel id")
		return
	}

	if err := h.users.UnsubscribeMonitorFromChannel(r.Context(), monitorID, channelID, userID); err != nil {
		switch {
		case errors.Is(err, services.ErrMonitorNotFound):
			Error(w, http.StatusNotFound, "monitor not found")
		case errors.Is(err, services.ErrAlertChannelNotFound):
			Error(w, http.StatusNotFound, "alert channel not found")
		default:
			h.log.ErrorContext(r.Context(), "unsubscribe monitor failed",
				"request_id", middleware.RequestIDFromContext(r.Context()),
				"user_id", userID,
				"monitor_id", monitorID,
				"channel_id", channelID,
				"error", err,
			)
			Error(w, http.StatusInternalServerError, "failed to unsubscribe")
		}
		return
	}

	h.log.InfoContext(r.Context(), "monitor unsubscribed from channel",
		"request_id", middleware.RequestIDFromContext(r.Context()),
		"user_id", userID,
		"monitor_id", monitorID,
		"channel_id", channelID,
	)
	w.WriteHeader(http.StatusNoContent)
}

// AvatarUploadURL returns a short-lived presigned PUT URL so the browser can
// upload an image directly to object storage without passing through the API.
//
// Request:  POST /api/v1/me/avatar-upload-url
//
//	body: { "content_type": "image/jpeg" }
//
// Response: { "upload_url": "...", "public_url": "..." }
//
// After the browser finishes the PUT, it calls PATCH /api/v1/me/avatar with
// public_url so the API can save it to the database.
func (h *UserHandler) AvatarUploadURL(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var body struct {
		ContentType string `json:"content_type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.ContentType == "" {
		Error(w, http.StatusBadRequest, "content_type is required")
		return
	}

	result, err := h.users.AvatarUploadURL(r.Context(), userID, body.ContentType)
	if err != nil {
		if errors.Is(err, services.ErrStorageNotConfigured) {
			Error(w, http.StatusNotImplemented, "avatar uploads not configured")
			return
		}
		h.log.ErrorContext(r.Context(), "avatar upload url failed",
			"request_id", middleware.RequestIDFromContext(r.Context()),
			"user_id", userID,
			"error", err,
		)
		Error(w, http.StatusBadRequest, err.Error())
		return
	}

	JSON(w, http.StatusOK, result)
}

// UpdateAvatar saves the public URL of an already-uploaded avatar to the database.
//
// Request:  PATCH /api/v1/me/avatar
//
//	body: { "avatar_url": "https://..." }
func (h *UserHandler) UpdateAvatar(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var body struct {
		AvatarURL string `json:"avatar_url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.AvatarURL == "" {
		Error(w, http.StatusBadRequest, "avatar_url is required")
		return
	}

	if err := h.users.UpdateAvatar(r.Context(), userID, body.AvatarURL); err != nil {
		h.log.ErrorContext(r.Context(), "update avatar failed",
			"request_id", middleware.RequestIDFromContext(r.Context()),
			"user_id", userID,
			"error", err,
		)
		Error(w, http.StatusInternalServerError, "failed to update avatar")
		return
	}

	h.log.InfoContext(r.Context(), "avatar updated",
		"request_id", middleware.RequestIDFromContext(r.Context()),
		"user_id", userID,
	)
	w.WriteHeader(http.StatusNoContent)
}

func (h *UserHandler) SubscribeMonitorToChannel(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	monitorID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid monitor id")
		return
	}

	var body struct {
		AlertChannelID uuid.UUID `json:"alert_channel_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.users.SubscribeMonitorToChannel(r.Context(), monitorID, body.AlertChannelID, userID); err != nil {
		switch {
		case errors.Is(err, services.ErrMonitorNotFound):
			Error(w, http.StatusNotFound, "monitor not found")
		case errors.Is(err, services.ErrAlertChannelNotFound):
			Error(w, http.StatusNotFound, "alert channel not found")
		default:
			h.log.ErrorContext(r.Context(), "subscribe monitor failed",
				"request_id", middleware.RequestIDFromContext(r.Context()),
				"user_id", userID,
				"monitor_id", monitorID,
				"channel_id", body.AlertChannelID,
				"error", err,
			)
			Error(w, http.StatusInternalServerError, "failed to subscribe")
		}
		return
	}

	h.log.InfoContext(r.Context(), "monitor subscribed to channel",
		"request_id", middleware.RequestIDFromContext(r.Context()),
		"user_id", userID,
		"monitor_id", monitorID,
		"channel_id", body.AlertChannelID,
	)
	JSON(w, http.StatusCreated, map[string]string{"message": "subscribed"})
}

func (h *UserHandler) UpdateMonitorMeta(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid monitor id")
		return
	}
	var body struct {
		Name        string  `json:"name"`
		Description string  `json:"description"`
		ComponentID *string `json:"component_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	var componentID *uuid.UUID
	if body.ComponentID != nil && *body.ComponentID != "" {
		parsed, err := uuid.Parse(*body.ComponentID)
		if err != nil {
			Error(w, http.StatusBadRequest, "invalid component_id")
			return
		}
		componentID = &parsed
	}
	monitor, err := h.users.UpdateMonitorMeta(r.Context(), id, userID, body.Name, body.Description, componentID)
	if err != nil {
		Error(w, http.StatusInternalServerError, "failed to update monitor")
		return
	}
	JSON(w, http.StatusOK, monitor)
}

// ── Status page slug ──────────────────────────────────────────────────────────

func (h *UserHandler) SetSlug(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var body struct {
		Slug string `json:"slug"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Slug == "" {
		Error(w, http.StatusBadRequest, "slug is required")
		return
	}

	if err := h.users.SetStatusPageSlug(r.Context(), userID, body.Slug); err != nil {
		if errors.Is(err, services.ErrSlugTaken) {
			Error(w, http.StatusConflict, "this URL is already taken")
			return
		}
		Error(w, http.StatusBadRequest, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ── Import alert channels ─────────────────────────────────────────────────────

func (h *UserHandler) ImportAlertChannels(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var body struct {
		Channels   []inbound.ImportChannelRow `json:"channels"`
		OnConflict string                     `json:"on_conflict"` // "skip" | "overwrite"
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if len(body.Channels) == 0 {
		Error(w, http.StatusBadRequest, "no channels provided")
		return
	}
	if len(body.Channels) > 500 {
		Error(w, http.StatusBadRequest, "too many channels (max 500 per import)")
		return
	}
	if body.OnConflict != "skip" && body.OnConflict != "overwrite" {
		body.OnConflict = "skip"
	}

	result, err := h.users.ImportAlertChannels(r.Context(), userID, body.Channels, body.OnConflict)
	if err != nil {
		h.log.ErrorContext(r.Context(), "import alert channels failed", "user_id", userID, "error", err)
		Error(w, http.StatusInternalServerError, "import failed")
		return
	}

	JSON(w, http.StatusOK, result)
}
