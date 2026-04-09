package checker

import (
	"context"
	"log"
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
	concLimit int // max concurrent checks
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

// Run starts the worker loop. It polls every 10 seconds for due monitors.
// Blocks until ctx is cancelled.
func (w *Worker) Run(ctx context.Context) {
	log.Printf("worker[%s] started (concurrency=%d)", w.region, w.concLimit)
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	// Run once immediately on start
	w.runBatch(ctx)

	for {
		select {
		case <-ctx.Done():
			log.Printf("worker[%s] shutting down", w.region)
			return
		case <-ticker.C:
			w.runBatch(ctx)
		}
	}
}

func (w *Worker) runBatch(ctx context.Context) {
	due, err := w.monitors.GetDue(ctx, w.region)
	if err != nil {
		log.Printf("worker[%s] get due monitors: %v", w.region, err)
		return
	}
	if len(due) == 0 {
		return
	}

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
		log.Printf("worker[%s] no checker for type %s", w.region, monitor.Type)
		return
	}

	result, err := checker.Check(ctx, monitor)
	if err != nil {
		log.Printf("worker[%s] check error monitor=%s: %v", w.region, monitor.ID, err)
		return
	}

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
		log.Printf("worker[%s] save check: %v", w.region, err)
		return
	}

	// Update monitor's last checked time and status
	monitor.LastCheckedAt = &now
	monitor.UpdatedAt = now

	previousStatus := monitor.Status

	if result.IsUp {
		monitor.Status = domain.MonitorStatusUp
	} else {
		monitor.Status = domain.MonitorStatusDown
	}

	if err := w.monitors.Update(ctx, &monitor); err != nil {
		log.Printf("worker[%s] update monitor: %v", w.region, err)
	}

	w.handleIncident(ctx, monitor, previousStatus, result.IsUp, now)
}

func (w *Worker) handleIncident(ctx context.Context, monitor domain.Monitor, previousStatus domain.MonitorStatus, isUp bool, now time.Time) {
	if !isUp && previousStatus != domain.MonitorStatusDown {
		// Transition to DOWN — open incident and alert
		incident := &domain.Incident{
			ID:        uuid.New(),
			MonitorID: monitor.ID,
			StartedAt: now,
		}
		if err := w.incidents.Create(ctx, incident); err != nil {
			log.Printf("worker[%s] create incident: %v", w.region, err)
			return
		}
		w.sendAlerts(ctx, domain.AlertEvent{
			Monitor:  monitor,
			Incident: *incident,
			Type:     domain.AlertEventDown,
		})
		return
	}

	if isUp && previousStatus == domain.MonitorStatusDown {
		// Recovery — resolve open incident and alert
		incident, err := w.incidents.GetOpenByMonitorID(ctx, monitor.ID)
		if err != nil {
			return
		}
		if err := w.incidents.Resolve(ctx, incident.ID, now); err != nil {
			log.Printf("worker[%s] resolve incident: %v", w.region, err)
			return
		}
		incident.ResolvedAt = &now
		w.sendAlerts(ctx, domain.AlertEvent{
			Monitor:  monitor,
			Incident: *incident,
			Type:     domain.AlertEventRecovery,
		})
	}
}

func (w *Worker) sendAlerts(ctx context.Context, event domain.AlertEvent) {
	channels, err := w.channels.GetByMonitorID(ctx, event.Monitor.ID)
	if err != nil {
		log.Printf("worker[%s] get alert channels: %v", w.region, err)
		return
	}

	for _, ch := range channels {
		notifier, ok := w.notifiers[ch.Type]
		if !ok {
			log.Printf("worker[%s] no notifier for type %s", w.region, ch.Type)
			continue
		}
		if err := notifier.Send(ctx, event, ch.Config); err != nil {
			log.Printf("worker[%s] send alert channel=%s: %v", w.region, ch.ID, err)
		}
	}
}
