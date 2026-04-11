// Package webhook provides Slack and Discord alert notifiers.
// Both platforms accept a simple HTTP POST with JSON — called an "incoming webhook".
// The user creates a webhook URL in Slack/Discord and pastes it into Pingr.
// When a monitor goes down, Pingr POSTs to that URL.
//
// Interview explanation:
//   - Slack and Discord both use the same pattern (incoming webhooks) but
//     different JSON shapes. We implement one notifier per platform.
//   - Adding Telegram or PagerDuty later = one new file, zero other changes.
package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/smaranbhupathi/pingr/internal/core/domain"
	"github.com/smaranbhupathi/pingr/internal/core/ports/outbound"
)

// ── Slack ─────────────────────────────────────────────────────────────────────

type slackNotifier struct{ client *http.Client }

func NewSlackNotifier() outbound.Notifier {
	return &slackNotifier{client: &http.Client{Timeout: 10 * time.Second}}
}

func (n *slackNotifier) Type() domain.AlertChannelType {
	return domain.AlertChannelSlack
}

func (n *slackNotifier) SendSubscriptionConfirmation(ctx context.Context, monitorName, monitorURL string, config map[string]any) error {
	webhookURL, ok := config["webhook_url"].(string)
	if !ok || webhookURL == "" {
		return fmt.Errorf("slack notifier: missing webhook_url in config")
	}
	text := fmt.Sprintf("✅ *%s* is now subscribed to Pingr alerts for *%s* (<%s|%s>)\nYou'll be notified here when the monitor goes down or recovers.",
		"This channel", monitorName, monitorURL, monitorURL,
	)
	return post(ctx, n.client, webhookURL, map[string]string{"text": text})
}

func (n *slackNotifier) Send(ctx context.Context, event domain.AlertEvent, config map[string]any) error {
	webhookURL, ok := config["webhook_url"].(string)
	if !ok || webhookURL == "" {
		return fmt.Errorf("slack notifier: missing webhook_url in config")
	}

	var text string
	switch event.Type {
	case domain.AlertEventDown:
		text = fmt.Sprintf("🔴 *%s* is DOWN\n<%s|%s>\nIncident started at %s UTC",
			event.Monitor.Name,
			event.Monitor.URL,
			event.Monitor.URL,
			event.OutageEvent.StartedAt.UTC().Format("2006-01-02 15:04:05"),
		)
	case domain.AlertEventRecovery:
		text = fmt.Sprintf("🟢 *%s* is back UP\n<%s|%s>\nRecovered at %s UTC",
			event.Monitor.Name,
			event.Monitor.URL,
			event.Monitor.URL,
			event.OutageEvent.ResolvedAt.UTC().Format("2006-01-02 15:04:05"),
		)
	}

	payload := map[string]string{"text": text}
	return post(ctx, n.client, webhookURL, payload)
}

// ── Discord ───────────────────────────────────────────────────────────────────

type discordNotifier struct{ client *http.Client }

func NewDiscordNotifier() outbound.Notifier {
	return &discordNotifier{client: &http.Client{Timeout: 10 * time.Second}}
}

func (n *discordNotifier) Type() domain.AlertChannelType {
	return domain.AlertChannelDiscord
}

func (n *discordNotifier) SendSubscriptionConfirmation(ctx context.Context, monitorName, monitorURL string, config map[string]any) error {
	webhookURL, ok := config["webhook_url"].(string)
	if !ok || webhookURL == "" {
		return fmt.Errorf("discord notifier: missing webhook_url in config")
	}
	content := fmt.Sprintf("✅ **%s** is now subscribed to Pingr alerts for **%s** (%s)\nYou'll be notified here when the monitor goes down or recovers.",
		"This channel", monitorName, monitorURL,
	)
	return post(ctx, n.client, webhookURL, map[string]string{"content": content})
}

func (n *discordNotifier) Send(ctx context.Context, event domain.AlertEvent, config map[string]any) error {
	webhookURL, ok := config["webhook_url"].(string)
	if !ok || webhookURL == "" {
		return fmt.Errorf("discord notifier: missing webhook_url in config")
	}

	var content string
	switch event.Type {
	case domain.AlertEventDown:
		content = fmt.Sprintf("🔴 **%s** is DOWN\n%s\nIncident started at %s UTC",
			event.Monitor.Name,
			event.Monitor.URL,
			event.OutageEvent.StartedAt.UTC().Format("2006-01-02 15:04:05"),
		)
	case domain.AlertEventRecovery:
		content = fmt.Sprintf("🟢 **%s** is back UP\n%s\nRecovered at %s UTC",
			event.Monitor.Name,
			event.Monitor.URL,
			event.OutageEvent.ResolvedAt.UTC().Format("2006-01-02 15:04:05"),
		)
	}

	payload := map[string]string{"content": content}
	return post(ctx, n.client, webhookURL, payload)
}

func (n *slackNotifier) SendIncidentUpdate(ctx context.Context, incident domain.Incident, update domain.IncidentUpdate, config map[string]any) error {
	webhookURL, ok := config["webhook_url"].(string)
	if !ok || webhookURL == "" {
		return fmt.Errorf("slack notifier: missing webhook_url in config")
	}

	emoji := incidentEmoji(update.Status)
	monitorNames := incidentMonitorNames(incident)

	text := fmt.Sprintf("%s *[%s] %s*\n\n%s\n\n*Affected:* %s\n*Updated:* %s UTC",
		emoji,
		incidentStatusLabel(update.Status),
		incident.Name,
		update.Message,
		monitorNames,
		update.CreatedAt.UTC().Format("2006-01-02 15:04:05"),
	)
	return post(ctx, n.client, webhookURL, map[string]string{"text": text})
}

func (n *discordNotifier) SendIncidentUpdate(ctx context.Context, incident domain.Incident, update domain.IncidentUpdate, config map[string]any) error {
	webhookURL, ok := config["webhook_url"].(string)
	if !ok || webhookURL == "" {
		return fmt.Errorf("discord notifier: missing webhook_url in config")
	}

	type field struct {
		Name   string `json:"name"`
		Value  string `json:"value"`
		Inline bool   `json:"inline"`
	}
	type embed struct {
		Title       string  `json:"title"`
		Description string  `json:"description"`
		Color       int     `json:"color"`
		Fields      []field `json:"fields"`
	}

	payload := map[string]any{
		"embeds": []embed{{
			Title:       fmt.Sprintf("%s Incident Update: %s", incidentEmoji(update.Status), incident.Name),
			Description: update.Message,
			Color:       incidentColor(update.Status),
			Fields: []field{
				{Name: "Status", Value: incidentStatusLabel(update.Status), Inline: true},
				{Name: "Affected", Value: incidentMonitorNames(incident), Inline: true},
				{Name: "Updated", Value: update.CreatedAt.UTC().Format("2006-01-02 15:04:05") + " UTC", Inline: false},
			},
		}},
	}
	return post(ctx, n.client, webhookURL, payload)
}

// ── helpers ───────────────────────────────────────────────────────────────────

func incidentStatusLabel(s domain.IncidentStatus) string {
	switch s {
	case domain.IncidentStatusInvestigating:
		return "Investigating"
	case domain.IncidentStatusIdentified:
		return "Identified"
	case domain.IncidentStatusMonitoring:
		return "Monitoring"
	case domain.IncidentStatusResolved:
		return "Resolved"
	default:
		return string(s)
	}
}

func incidentEmoji(s domain.IncidentStatus) string {
	switch s {
	case domain.IncidentStatusInvestigating:
		return "🔴"
	case domain.IncidentStatusIdentified:
		return "🟠"
	case domain.IncidentStatusMonitoring:
		return "🟡"
	case domain.IncidentStatusResolved:
		return "🟢"
	default:
		return "⚪"
	}
}

func incidentColor(s domain.IncidentStatus) int {
	switch s {
	case domain.IncidentStatusInvestigating:
		return 0xdc2626 // red
	case domain.IncidentStatusIdentified:
		return 0xf97316 // orange
	case domain.IncidentStatusMonitoring:
		return 0xeab308 // yellow
	case domain.IncidentStatusResolved:
		return 0x16a34a // green
	default:
		return 0x6b7280
	}
}

func incidentMonitorNames(incident domain.Incident) string {
	if len(incident.Monitors) == 0 {
		return "—"
	}
	names := make([]string, 0, len(incident.Monitors))
	for _, m := range incident.Monitors {
		names = append(names, m.Name)
	}
	result := ""
	for i, n := range names {
		if i > 0 {
			result += ", "
		}
		result += n
	}
	return result
}

// ── shared ────────────────────────────────────────────────────────────────────

func post(ctx context.Context, client *http.Client, url string, payload any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("webhook post: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook returned %d", resp.StatusCode)
	}
	return nil
}
