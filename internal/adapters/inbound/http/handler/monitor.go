package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
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
	log      *slog.Logger
}

func NewMonitorHandler(monitors inbound.MonitorService, log *slog.Logger) *MonitorHandler {
	return &MonitorHandler{monitors: monitors, log: log}
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
		body.IntervalSeconds = 60
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
			h.log.ErrorContext(r.Context(), "create monitor failed",
				"request_id", middleware.RequestIDFromContext(r.Context()),
				"user_id", userID,
				"url", body.URL,
				"error", err,
			)
			Error(w, http.StatusInternalServerError, "failed to create monitor")
		}
		return
	}

	h.log.InfoContext(r.Context(), "monitor created",
		"request_id", middleware.RequestIDFromContext(r.Context()),
		"user_id", userID,
		"monitor_id", monitor.ID,
		"url", monitor.URL,
	)
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
		h.log.ErrorContext(r.Context(), "list monitors failed",
			"request_id", middleware.RequestIDFromContext(r.Context()),
			"user_id", userID,
			"error", err,
		)
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
		h.log.ErrorContext(r.Context(), "get monitor failed",
			"request_id", middleware.RequestIDFromContext(r.Context()),
			"user_id", userID,
			"monitor_id", id,
			"error", err,
		)
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
			h.log.ErrorContext(r.Context(), "update monitor failed",
				"request_id", middleware.RequestIDFromContext(r.Context()),
				"user_id", userID,
				"monitor_id", id,
				"error", err,
			)
			Error(w, http.StatusInternalServerError, "failed to update monitor")
		}
		return
	}

	h.log.InfoContext(r.Context(), "monitor updated",
		"request_id", middleware.RequestIDFromContext(r.Context()),
		"user_id", userID,
		"monitor_id", monitor.ID,
		"is_active", monitor.IsActive,
		"interval_seconds", monitor.IntervalSeconds,
	)
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
		h.log.ErrorContext(r.Context(), "delete monitor failed",
			"request_id", middleware.RequestIDFromContext(r.Context()),
			"user_id", userID,
			"monitor_id", id,
			"error", err,
		)
		Error(w, http.StatusInternalServerError, "failed to delete monitor")
		return
	}

	h.log.InfoContext(r.Context(), "monitor deleted",
		"request_id", middleware.RequestIDFromContext(r.Context()),
		"user_id", userID,
		"monitor_id", id,
	)
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
		h.log.ErrorContext(r.Context(), "get graph failed",
			"request_id", middleware.RequestIDFromContext(r.Context()),
			"monitor_id", id,
			"error", err,
		)
		Error(w, http.StatusInternalServerError, "failed to get graph data")
		return
	}

	JSON(w, http.StatusOK, points)
}

func (h *MonitorHandler) StatusPage(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")

	details, err := h.monitors.GetStatusPage(r.Context(), username)
	if err != nil {
		h.log.ErrorContext(r.Context(), "status page failed",
			"request_id", middleware.RequestIDFromContext(r.Context()),
			"username", username,
			"error", err,
		)
		Error(w, http.StatusInternalServerError, "failed to load status page")
		return
	}

	JSON(w, http.StatusOK, details)
}
