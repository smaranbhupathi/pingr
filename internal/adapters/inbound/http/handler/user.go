package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/smaranbhupathi/pingr/internal/adapters/inbound/http/middleware"
	"github.com/smaranbhupathi/pingr/internal/core/domain"
	"github.com/smaranbhupathi/pingr/internal/core/ports/inbound"
	"github.com/smaranbhupathi/pingr/internal/core/services"
)

type UserHandler struct {
	users inbound.UserService
}

func NewUserHandler(users inbound.UserService) *UserHandler {
	return &UserHandler{users: users}
}

func (h *UserHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	profile, err := h.users.GetProfile(r.Context(), userID)
	if err != nil {
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
		Type      domain.AlertChannelType `json:"type"`
		Config    map[string]any          `json:"config"`
		IsDefault bool                    `json:"is_default"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	input := inbound.CreateAlertChannelInput{
		UserID:    userID,
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
		Error(w, http.StatusInternalServerError, "failed to create alert channel")
		return
	}

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
		Error(w, http.StatusInternalServerError, "failed to list alert channels")
		return
	}

	JSON(w, http.StatusOK, channels)
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
		Error(w, http.StatusInternalServerError, "failed to delete alert channel")
		return
	}

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
			Error(w, http.StatusInternalServerError, "failed to subscribe")
		}
		return
	}

	JSON(w, http.StatusCreated, map[string]string{"message": "subscribed"})
}
