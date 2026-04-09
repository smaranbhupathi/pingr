package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/smaranbhupathi/pingr/internal/adapters/inbound/http/middleware"
	"github.com/smaranbhupathi/pingr/internal/core/ports/inbound"
	"github.com/smaranbhupathi/pingr/internal/core/services"
)

type MonitorHandler struct {
	monitors inbound.MonitorService
}

func NewMonitorHandler(monitors inbound.MonitorService) *MonitorHandler {
	return &MonitorHandler{monitors: monitors}
}

func (h *MonitorHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var body struct {
		Name            string `json:"name"`
		URL             string `json:"url"`
		IntervalSeconds int    `json:"interval_seconds"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if body.IntervalSeconds == 0 {
		body.IntervalSeconds = 60 // sensible default
	}

	createInput := inbound.CreateMonitorInput{
		UserID:          userID,
		Name:            body.Name,
		URL:             body.URL,
		IntervalSeconds: body.IntervalSeconds,
	}
	if errs := createInput.Validate(); len(errs) > 0 {
		JSON(w, http.StatusUnprocessableEntity, map[string]any{"errors": errs})
		return
	}

	monitor, err := h.monitors.Create(r.Context(), createInput)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrMonitorLimitReached):
			Error(w, http.StatusForbidden, "monitor limit reached for your plan")
		case errors.Is(err, services.ErrInvalidInterval):
			Error(w, http.StatusBadRequest, "check interval below plan minimum")
		default:
			Error(w, http.StatusInternalServerError, "failed to create monitor")
		}
		return
	}

	JSON(w, http.StatusCreated, monitor)
}

func (h *MonitorHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	monitors, err := h.monitors.ListByUser(r.Context(), userID)
	if err != nil {
		Error(w, http.StatusInternalServerError, "failed to list monitors")
		return
	}

	JSON(w, http.StatusOK, monitors)
}

func (h *MonitorHandler) Get(w http.ResponseWriter, r *http.Request) {
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

	detail, err := h.monitors.GetByID(r.Context(), id, userID)
	if err != nil {
		if errors.Is(err, services.ErrMonitorNotFound) {
			Error(w, http.StatusNotFound, "monitor not found")
			return
		}
		Error(w, http.StatusInternalServerError, "failed to get monitor")
		return
	}

	JSON(w, http.StatusOK, detail)
}

func (h *MonitorHandler) Update(w http.ResponseWriter, r *http.Request) {
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

	var input inbound.UpdateMonitorInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	monitor, err := h.monitors.Update(r.Context(), id, userID, input)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrMonitorNotFound):
			Error(w, http.StatusNotFound, "monitor not found")
		case errors.Is(err, services.ErrInvalidInterval):
			Error(w, http.StatusBadRequest, "check interval below plan minimum")
		default:
			Error(w, http.StatusInternalServerError, "failed to update monitor")
		}
		return
	}

	JSON(w, http.StatusOK, monitor)
}

func (h *MonitorHandler) Delete(w http.ResponseWriter, r *http.Request) {
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

	if err := h.monitors.Delete(r.Context(), id, userID); err != nil {
		if errors.Is(err, services.ErrMonitorNotFound) {
			Error(w, http.StatusNotFound, "monitor not found")
			return
		}
		Error(w, http.StatusInternalServerError, "failed to delete monitor")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *MonitorHandler) ResponseTimeGraph(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid monitor id")
		return
	}

	from := time.Now().Add(-24 * time.Hour)
	to := time.Now()

	points, err := h.monitors.GetResponseTimeGraph(r.Context(), id, from, to)
	if err != nil {
		Error(w, http.StatusInternalServerError, "failed to get graph data")
		return
	}

	JSON(w, http.StatusOK, points)
}

// StatusPage is public — no auth required.
func (h *MonitorHandler) StatusPage(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")

	details, err := h.monitors.GetStatusPage(r.Context(), username)
	if err != nil {
		Error(w, http.StatusInternalServerError, "failed to load status page")
		return
	}

	JSON(w, http.StatusOK, details)
}
