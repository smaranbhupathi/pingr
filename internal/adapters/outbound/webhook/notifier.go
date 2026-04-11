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
			event.Incident.StartedAt.UTC().Format("2006-01-02 15:04:05"),
		)
	case domain.AlertEventRecovery:
		text = fmt.Sprintf("🟢 *%s* is back UP\n<%s|%s>\nRecovered at %s UTC",
			event.Monitor.Name,
			event.Monitor.URL,
			event.Monitor.URL,
			event.Incident.ResolvedAt.UTC().Format("2006-01-02 15:04:05"),
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
			event.Incident.StartedAt.UTC().Format("2006-01-02 15:04:05"),
		)
	case domain.AlertEventRecovery:
		content = fmt.Sprintf("🟢 **%s** is back UP\n%s\nRecovered at %s UTC",
			event.Monitor.Name,
			event.Monitor.URL,
			event.Incident.ResolvedAt.UTC().Format("2006-01-02 15:04:05"),
		)
	}

	payload := map[string]string{"content": content}
	return post(ctx, n.client, webhookURL, payload)
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
