package checker

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/smaranbhupathi/pingr/internal/core/domain"
	"github.com/smaranbhupathi/pingr/internal/core/ports/outbound"
)

// Worker polls the database for monitors due for checking and runs them concurrently.
// Each worker is scoped to a region — scaling to multi-region means deploying more workers
// with different region tags. No code changes needed.
type Worker struct {
	region    string
	monitors  outbound.MonitorRepository
	checks    outbound.CheckRepository
	incidents outbound.IncidentRepository
	channels  outbound.AlertChannelRepository
	checkers  map[domain.MonitorType]outbound.Checker
	notifiers map[domain.AlertChannelType]outbound.Notifier
	concLimit int
}

func NewWorker(
	region string,
	monitors outbound.MonitorRepository,
	checks outbound.CheckRepository,
	incidents outbound.IncidentRepository,
	channels outbound.AlertChannelRepository,
	checkers []outbound.Checker,
	notifiers []outbound.Notifier,
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
		region:    region,
		monitors:  monitors,
		checks:    checks,
		incidents: incidents,
		channels:  channels,
		checkers:  checkerMap,
		notifiers: notifierMap,
		concLimit: concLimit,
	}
}

// Run starts the worker loop. Polls every 10 seconds for due monitors.
func (w *Worker) Run(ctx context.Context) {
	slog.Info("worker started", "region", w.region, "concurrency", w.concLimit)
	ticker := time.NewTicker(10 * time.Second)
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
		incident := &domain.Incident{
			ID:        uuid.New(),
			MonitorID: monitor.ID,
			StartedAt: now,
		}
		if err := w.incidents.Create(ctx, incident); err != nil {
			slog.Error("create incident failed", "region", w.region, "monitor_id", monitor.ID, "error", err)
			return
		}
		slog.Warn("monitor DOWN — incident opened",
			"region", w.region,
			"monitor_id", monitor.ID,
			"monitor_name", monitor.Name,
			"url", monitor.URL,
			"incident_id", incident.ID,
		)
		w.sendAlerts(ctx, domain.AlertEvent{Monitor: monitor, Incident: *incident, Type: domain.AlertEventDown})
		return
	}

	if isUp && previousStatus == domain.MonitorStatusDown {
		incident, err := w.incidents.GetOpenByMonitorID(ctx, monitor.ID)
		if err != nil {
			return
		}
		if err := w.incidents.Resolve(ctx, incident.ID, now); err != nil {
			slog.Error("resolve incident failed", "region", w.region, "monitor_id", monitor.ID, "error", err)
			return
		}
		incident.ResolvedAt = &now
		slog.Info("monitor RECOVERED — incident resolved",
			"region", w.region,
			"monitor_id", monitor.ID,
			"monitor_name", monitor.Name,
			"url", monitor.URL,
			"incident_id", incident.ID,
		)
		w.sendAlerts(ctx, domain.AlertEvent{Monitor: monitor, Incident: *incident, Type: domain.AlertEventRecovery})
	}
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
