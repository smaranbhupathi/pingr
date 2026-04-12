package checker

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/smaranbhupathi/pingr/internal/config"
	"github.com/smaranbhupathi/pingr/internal/core/domain"
	"github.com/smaranbhupathi/pingr/internal/core/ports/outbound"
)

// Worker polls the database for monitors due for checking and runs them concurrently.
// Each worker is scoped to a region — scaling to multi-region means deploying more workers
// with different region tags. No code changes needed.
type Worker struct {
	region       string
	monitors     outbound.MonitorRepository
	checks       outbound.CheckRepository
	outageEvents outbound.OutageEventRepository
	incidents    outbound.IncidentRepository
	channels     outbound.AlertChannelRepository
	checkers     map[domain.MonitorType]outbound.Checker
	notifiers    map[domain.AlertChannelType]outbound.Notifier
	cfg          *config.Config
	concLimit    int
}

func NewWorker(
	region string,
	monitors outbound.MonitorRepository,
	checks outbound.CheckRepository,
	outageEvents outbound.OutageEventRepository,
	incidents outbound.IncidentRepository,
	channels outbound.AlertChannelRepository,
	checkers []outbound.Checker,
	notifiers []outbound.Notifier,
	cfg *config.Config,
	concLimit int,
) *Worker {
	checkerMap := make(map[domain.MonitorType]outbound.Checker)
	for _, c := range checkers {
		checkerMap[c.Type()] = c
	}
	notifierMap := make(map[domain.AlertChannelType]outbound.Notifier)
	for _, n := range notifiers {
		notifierMap[n.Type()] = n
	}
	return &Worker{
		region:       region,
		monitors:     monitors,
		checks:       checks,
		outageEvents: outageEvents,
		incidents:    incidents,
		channels:     channels,
		checkers:     checkerMap,
		notifiers:    notifierMap,
		cfg:          cfg,
		concLimit:    concLimit,
	}
}

// Run starts the worker loop. Polls every 10 seconds for due monitors.
func (w *Worker) Run(ctx context.Context) {
	tick := time.Duration(w.cfg.Monitoring.WorkerTickSeconds) * time.Second
	slog.Info("worker started", "region", w.region, "concurrency", w.concLimit, "tick", tick)
	ticker := time.NewTicker(tick)
	defer ticker.Stop()

	w.runBatch(ctx)

	for {
		select {
		case <-ctx.Done():
			slog.Info("worker shutting down", "region", w.region)
			return
		case <-ticker.C:
			w.runBatch(ctx)
		}
	}
}

func (w *Worker) runBatch(ctx context.Context) {
	due, err := w.monitors.GetDue(ctx, w.region)
	if err != nil {
		slog.Error("get due monitors failed", "region", w.region, "error", err)
		return
	}
	if len(due) == 0 {
		return
	}

	slog.Debug("running batch", "region", w.region, "count", len(due))

	sem := make(chan struct{}, w.concLimit)
	var wg sync.WaitGroup

	for _, monitor := range due {
		select {
		case <-ctx.Done():
			return
		default:
		}
		sem <- struct{}{}
		wg.Add(1)
		go func(m domain.Monitor) {
			defer wg.Done()
			defer func() { <-sem }()
			w.checkMonitor(ctx, m)
		}(monitor)
	}
	wg.Wait()
}

func (w *Worker) checkMonitor(ctx context.Context, monitor domain.Monitor) {
	checker, ok := w.checkers[monitor.Type]
	if !ok {
		slog.Warn("no checker for monitor type", "region", w.region, "type", monitor.Type, "monitor_id", monitor.ID)
		return
	}

	start := time.Now()
	result, err := checker.Check(ctx, monitor)
	if err != nil {
		slog.Error("check failed", "region", w.region, "monitor_id", monitor.ID, "url", monitor.URL, "error", err)
		return
	}

	latency := time.Since(start).Milliseconds()
	slog.Debug("check complete",
		"region", w.region,
		"monitor_id", monitor.ID,
		"url", monitor.URL,
		"is_up", result.IsUp,
		"status_code", result.StatusCode,
		"latency_ms", latency,
	)

	now := time.Now()
	check := &domain.MonitorCheck{
		ID:             uuid.New(),
		MonitorID:      monitor.ID,
		CheckedAt:      now,
		IsUp:           result.IsUp,
		StatusCode:     result.StatusCode,
		ResponseTimeMs: result.ResponseTimeMs,
		ErrorMessage:   result.ErrorMessage,
		Region:         w.region,
	}

	if err := w.checks.Create(ctx, check); err != nil {
		slog.Error("save check failed", "region", w.region, "monitor_id", monitor.ID, "error", err)
		return
	}

	monitor.LastCheckedAt = &now
	monitor.UpdatedAt = now
	previousStatus := monitor.Status

	if result.IsUp {
		monitor.Status = domain.MonitorStatusUp
	} else {
		monitor.Status = domain.MonitorStatusDown
	}

	if err := w.monitors.Update(ctx, &monitor); err != nil {
		slog.Error("update monitor status failed", "region", w.region, "monitor_id", monitor.ID, "error", err)
	}

	w.handleIncident(ctx, monitor, previousStatus, result.IsUp, now)
}

func (w *Worker) handleIncident(ctx context.Context, monitor domain.Monitor, previousStatus domain.MonitorStatus, isUp bool, now time.Time) {
	if !isUp && previousStatus != domain.MonitorStatusDown {
		// Set component status to major_outage on the public status page.
		if err := w.monitors.UpdateComponentStatus(ctx, monitor.ID, domain.ComponentStatusMajorOutage); err != nil {
			slog.Error("update component status failed", "monitor_id", monitor.ID, "error", err)
		}

		// Create internal outage event (uptime math + alert triggering).
		outageEvent := &domain.OutageEvent{
			ID:        uuid.New(),
			MonitorID: monitor.ID,
			StartedAt: now,
		}
		if err := w.outageEvents.Create(ctx, outageEvent); err != nil {
			slog.Error("create outage event failed", "region", w.region, "monitor_id", monitor.ID, "error", err)
			return
		}

		// Auto-create a user-facing incident linked to this outage event.
		w.autoCreateIncident(ctx, monitor, outageEvent.ID, now)

		slog.Warn("monitor DOWN — outage event opened",
			"region", w.region,
			"monitor_id", monitor.ID,
			"monitor_name", monitor.Name,
			"url", monitor.URL,
			"outage_event_id", outageEvent.ID,
		)
		w.sendAlerts(ctx, domain.AlertEvent{Monitor: monitor, OutageEvent: *outageEvent, Type: domain.AlertEventDown})
		return
	}

	if isUp && previousStatus == domain.MonitorStatusDown {
		// Restore component status to operational on recovery.
		if err := w.monitors.UpdateComponentStatus(ctx, monitor.ID, domain.ComponentStatusOperational); err != nil {
			slog.Error("update component status failed", "monitor_id", monitor.ID, "error", err)
		}

		outageEvent, err := w.outageEvents.GetOpenByMonitorID(ctx, monitor.ID)
		if err != nil {
			return
		}
		if err := w.outageEvents.Resolve(ctx, outageEvent.ID, now); err != nil {
			slog.Error("resolve outage event failed", "region", w.region, "monitor_id", monitor.ID, "error", err)
			return
		}
		outageEvent.ResolvedAt = &now

		// Auto-resolve the user-facing incident that belongs to this outage event.
		w.autoResolveIncident(ctx, outageEvent.ID, monitor.Name, now)

		slog.Info("monitor RECOVERED — outage event resolved",
			"region", w.region,
			"monitor_id", monitor.ID,
			"monitor_name", monitor.Name,
			"url", monitor.URL,
			"outage_event_id", outageEvent.ID,
		)
		w.sendAlerts(ctx, domain.AlertEvent{Monitor: monitor, OutageEvent: *outageEvent, Type: domain.AlertEventRecovery})
	}
}

func (w *Worker) autoCreateIncident(ctx context.Context, monitor domain.Monitor, outageEventID uuid.UUID, now time.Time) {
	inc := &domain.Incident{
		ID:            uuid.New(),
		UserID:        monitor.UserID,
		Name:          monitor.Name + " outage",
		Status:        domain.IncidentStatusInvestigating,
		Source:        "auto",
		OutageEventID: &outageEventID,
		MonitorIDs:    []uuid.UUID{monitor.ID},
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := w.incidents.Create(ctx, inc); err != nil {
		slog.Error("auto-create incident failed", "monitor_id", monitor.ID, "error", err)
		return
	}

	update := &domain.IncidentUpdate{
		ID:         uuid.New(),
		IncidentID: inc.ID,
		Status:     domain.IncidentStatusInvestigating,
		Message:    "We are investigating connectivity issues with " + monitor.Name + ".",
		Notify:     false,
		Source:     "auto",
		CreatedAt:  now,
	}
	if err := w.incidents.AddUpdate(ctx, update); err != nil {
		slog.Error("auto-create incident update failed", "incident_id", inc.ID, "error", err)
	}
}

func (w *Worker) autoResolveIncident(ctx context.Context, outageEventID uuid.UUID, monitorName string, now time.Time) {
	// Look up by outage_event_id — not by monitor_id.
	// This ensures we only touch the exact incident this outage spawned,
	// leaving any other open manual incidents for the same monitor untouched.
	inc, err := w.incidents.GetByOutageEventID(ctx, outageEventID)
	if err != nil {
		slog.Warn("auto-resolve: no incident linked to outage event (may have been manually resolved or never created)",
			"outage_event_id", outageEventID,
			"monitor_name", monitorName,
			"error", err,
		)
		return
	}
	if inc.ResolvedAt != nil {
		// Operator already resolved it manually — respect that, don't overwrite.
		slog.Debug("auto-resolve: incident already resolved manually, skipping", "incident_id", inc.ID)
		return
	}

	update := &domain.IncidentUpdate{
		ID:         uuid.New(),
		IncidentID: inc.ID,
		Status:     domain.IncidentStatusResolved,
		Message:    monitorName + " has recovered and is operating normally.",
		Notify:     false,
		Source:     "auto",
		CreatedAt:  now,
	}
	if err := w.incidents.AddUpdate(ctx, update); err != nil {
		slog.Error("auto-resolve incident update failed", "incident_id", inc.ID, "error", err)
	}
	if err := w.incidents.Resolve(ctx, inc.ID); err != nil {
		slog.Error("auto-resolve incident failed", "incident_id", inc.ID, "error", err)
		return
	}
	slog.Info("incident auto-resolved", "incident_id", inc.ID, "monitor_name", monitorName)
}

func (w *Worker) sendAlerts(ctx context.Context, event domain.AlertEvent) {
	channels, err := w.channels.GetByMonitorID(ctx, event.Monitor.ID)
	if err != nil {
		slog.Error("get alert channels failed", "monitor_id", event.Monitor.ID, "error", err)
		return
	}
	if len(channels) == 0 {
		slog.Debug("no alert channels for monitor", "monitor_id", event.Monitor.ID)
		return
	}

	for _, ch := range channels {
		// Per-channel toggle — user disabled this channel
		if !ch.IsEnabled {
			slog.Debug("alert channel disabled, skipping", "channel_id", ch.ID, "type", ch.Type)
			continue
		}

		// Platform-level feature flag — channel type turned off in config.yaml
		switch ch.Type {
		case domain.AlertChannelEmail:
			if !w.cfg.Features.EmailAlerts {
				slog.Debug("email alerts disabled in config, skipping", "channel_id", ch.ID)
				continue
			}
		case domain.AlertChannelSlack:
			if !w.cfg.Features.SlackAlerts {
				slog.Debug("slack alerts disabled in config, skipping", "channel_id", ch.ID)
				continue
			}
		case domain.AlertChannelDiscord:
			if !w.cfg.Features.DiscordAlerts {
				slog.Debug("discord alerts disabled in config, skipping", "channel_id", ch.ID)
				continue
			}
		}

		notifier, ok := w.notifiers[ch.Type]
		if !ok {
			slog.Warn("no notifier for channel type", "type", ch.Type, "channel_id", ch.ID)
			continue
		}
		if err := notifier.Send(ctx, event, ch.Config); err != nil {
			slog.Error("send alert failed", "channel_id", ch.ID, "type", ch.Type, "monitor_id", event.Monitor.ID, "error", err)
		} else {
			slog.Info("alert sent", "channel_id", ch.ID, "type", ch.Type, "monitor_id", event.Monitor.ID, "event_type", event.Type)
		}
	}
}
