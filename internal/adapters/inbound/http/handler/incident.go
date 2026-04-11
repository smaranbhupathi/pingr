package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/smaranbhupathi/pingr/internal/adapters/inbound/http/middleware"
	"github.com/smaranbhupathi/pingr/internal/core/domain"
	"github.com/smaranbhupathi/pingr/internal/core/ports/inbound"
	"github.com/smaranbhupathi/pingr/internal/core/services"
)

type IncidentHandler struct {
	users inbound.UserService
	log   *slog.Logger
}

func NewIncidentHandler(users inbound.UserService, log *slog.Logger) *IncidentHandler {
	return &IncidentHandler{users: users, log: log}
}

func (h *IncidentHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	incidents, err := h.users.ListIncidents(r.Context(), userID)
	if err != nil {
		h.log.ErrorContext(r.Context(), "list incidents failed", "user_id", userID, "error", err)
		Error(w, http.StatusInternalServerError, "failed to list incidents")
		return
	}

	JSON(w, http.StatusOK, incidents)
}

func (h *IncidentHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var body struct {
		Name       string                 `json:"name"`
		Status     domain.IncidentStatus  `json:"status"`
		Message    string                 `json:"message"`
		MonitorIDs []string               `json:"monitor_ids"`
		Notify     bool                   `json:"notify"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.Name == "" {
		Error(w, http.StatusUnprocessableEntity, "name is required")
		return
	}
	if body.Message == "" {
		Error(w, http.StatusUnprocessableEntity, "message is required")
		return
	}
	if body.Status == "" {
		body.Status = domain.IncidentStatusInvestigating
	}

	monitorIDs := make([]uuid.UUID, 0, len(body.MonitorIDs))
	for _, s := range body.MonitorIDs {
		id, err := uuid.Parse(s)
		if err != nil {
			Error(w, http.StatusBadRequest, "invalid monitor_id: "+s)
			return
		}
		monitorIDs = append(monitorIDs, id)
	}

	inc, err := h.users.CreateIncident(r.Context(), inbound.CreateIncidentInput{
		UserID:     userID,
		Name:       body.Name,
		Status:     body.Status,
		Message:    body.Message,
		MonitorIDs: monitorIDs,
		Notify:     body.Notify,
	})
	if err != nil {
		h.log.ErrorContext(r.Context(), "create incident failed", "user_id", userID, "error", err)
		Error(w, http.StatusInternalServerError, "failed to create incident")
		return
	}

	JSON(w, http.StatusCreated, inc)
}

func (h *IncidentHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid incident id")
		return
	}

	inc, err := h.users.GetIncident(r.Context(), id, userID)
	if err != nil {
		if errors.Is(err, services.ErrIncidentNotFound) {
			Error(w, http.StatusNotFound, "incident not found")
			return
		}
		Error(w, http.StatusInternalServerError, "failed to get incident")
		return
	}

	JSON(w, http.StatusOK, inc)
}

func (h *IncidentHandler) PostUpdate(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid incident id")
		return
	}

	var body struct {
		Status  domain.IncidentStatus `json:"status"`
		Message string                `json:"message"`
		Notify  bool                  `json:"notify"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.Message == "" {
		Error(w, http.StatusUnprocessableEntity, "message is required")
		return
	}
	if body.Status == "" {
		Error(w, http.StatusUnprocessableEntity, "status is required")
		return
	}

	inc, err := h.users.PostIncidentUpdate(r.Context(), inbound.PostIncidentUpdateInput{
		IncidentID: id,
		UserID:     userID,
		Status:     body.Status,
		Message:    body.Message,
		Notify:     body.Notify,
	})
	if err != nil {
		if errors.Is(err, services.ErrIncidentNotFound) {
			Error(w, http.StatusNotFound, "incident not found")
			return
		}
		h.log.ErrorContext(r.Context(), "post incident update failed", "incident_id", id, "error", err)
		Error(w, http.StatusInternalServerError, "failed to post update")
		return
	}

	JSON(w, http.StatusOK, inc)
}
