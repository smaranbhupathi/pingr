// Package webhook provides Slack and Discord alert notifiers.
package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/smaranbhupathi/pingr/internal/core/domain"
	"github.com/smaranbhupathi/pingr/internal/core/ports/outbound"
)

// notification is the canonical shape every alert is normalised into before
// being rendered. Both monitor events and incident updates map into this struct
// so every platform has exactly one renderer and the output is always consistent.
type notification struct {
	Emoji    string // 🔴 🟠 🟡 🟢
	TypeTag  string // DOWN | RECOVERED | INVESTIGATING | IDENTIFIED | MONITORING | RESOLVED
	Name     string // monitor name  or  incident name
	Message  string // human-readable description
	Affected string // "My server (https://…)"  or  "Monitor1, Monitor2"
	Time     string // "2026-04-11 13:35:13 UTC"
	Color    int    // Discord embed colour (hex int)
}

// fromAlertEvent converts a monitor up/down event into a notification.
func fromAlertEvent(event domain.AlertEvent) notification {
	switch event.Type {
	case domain.AlertEventDown:
		return notification{
			Emoji:    "🔴",
			TypeTag:  "DOWN",
			Name:     event.Monitor.Name,
			Message:  fmt.Sprintf("%s is unreachable.", event.Monitor.Name),
			Affected: fmt.Sprintf("%s — %s", event.Monitor.Name, event.Monitor.URL),
			Time:     event.OutageEvent.StartedAt.UTC().Format("2006-01-02 15:04:05 UTC"),
			Color:    0xdc2626,
		}
	case domain.AlertEventRecovery:
		return notification{
			Emoji:    "🟢",
			TypeTag:  "RECOVERED",
			Name:     event.Monitor.Name,
			Message:  fmt.Sprintf("%s has recovered and is back online.", event.Monitor.Name),
			Affected: fmt.Sprintf("%s — %s", event.Monitor.Name, event.Monitor.URL),
			Time:     event.OutageEvent.ResolvedAt.UTC().Format("2006-01-02 15:04:05 UTC"),
			Color:    0x16a34a,
		}
	default:
		return notification{}
	}
}

// fromIncidentUpdate converts an incident update into a notification.
func fromIncidentUpdate(incident domain.Incident, update domain.IncidentUpdate) notification {
	affected := "—"
	if len(incident.Monitors) > 0 {
		names := make([]string, 0, len(incident.Monitors))
		for _, m := range incident.Monitors {
			names = append(names, m.Name)
		}
		affected = strings.Join(names, ", ")
	}

	emoji, typeTag, color := statusMeta(update.Status)
	return notification{
		Emoji:    emoji,
		TypeTag:  typeTag,
		Name:     incident.Name,
		Message:  update.Message,
		Affected: affected,
		Time:     update.CreatedAt.UTC().Format("2006-01-02 15:04:05 UTC"),
		Color:    color,
	}
}

func statusMeta(s domain.IncidentStatus) (emoji, typeTag string, color int) {
	switch s {
	case domain.IncidentStatusInvestigating:
		return "🔴", "INVESTIGATING", 0xdc2626
	case domain.IncidentStatusIdentified:
		return "🟠", "IDENTIFIED", 0xf97316
	case domain.IncidentStatusMonitoring:
		return "🟡", "MONITORING", 0xeab308
	case domain.IncidentStatusResolved:
		return "🟢", "RESOLVED", 0x16a34a
	default:
		return "⚪", strings.ToUpper(string(s)), 0x6b7280
	}
}

// ── Slack ─────────────────────────────────────────────────────────────────────

type slackNotifier struct{ client *http.Client }

func NewSlackNotifier() outbound.Notifier {
	return &slackNotifier{client: &http.Client{Timeout: 10 * time.Second}}
}

func (n *slackNotifier) Type() domain.AlertChannelType { return domain.AlertChannelSlack }

func (n *slackNotifier) SendSubscriptionConfirmation(ctx context.Context, monitorName, monitorURL string, config map[string]any) error {
	url, err := webhookURL(config, "slack")
	if err != nil {
		return err
	}
	text := fmt.Sprintf("✅ *This channel* is now subscribed to Pingr alerts for *%s* (<%s|%s>)\nYou'll be notified here when the monitor goes down or recovers.",
		monitorName, monitorURL, monitorURL,
	)
	return post(ctx, n.client, url, map[string]string{"text": text})
}

func (n *slackNotifier) Send(ctx context.Context, event domain.AlertEvent, config map[string]any) error {
	url, err := webhookURL(config, "slack")
	if err != nil {
		return err
	}
	return post(ctx, n.client, url, map[string]string{"text": slackText(fromAlertEvent(event))})
}

func (n *slackNotifier) SendIncidentUpdate(ctx context.Context, incident domain.Incident, update domain.IncidentUpdate, config map[string]any) error {
	url, err := webhookURL(config, "slack")
	if err != nil {
		return err
	}
	return post(ctx, n.client, url, map[string]string{"text": slackText(fromIncidentUpdate(incident, update))})
}

// slackText renders a notification into Slack mrkdwn text.
func slackText(n notification) string {
	return fmt.Sprintf(
		"%s *[%s]* %s\n\n%s\n\n*Affected:* %s\n*Time:* %s",
		n.Emoji, n.TypeTag, n.Name,
		n.Message,
		n.Affected,
		n.Time,
	)
}

// ── Discord ───────────────────────────────────────────────────────────────────

type discordNotifier struct{ client *http.Client }

func NewDiscordNotifier() outbound.Notifier {
	return &discordNotifier{client: &http.Client{Timeout: 10 * time.Second}}
}

func (n *discordNotifier) Type() domain.AlertChannelType { return domain.AlertChannelDiscord }

func (n *discordNotifier) SendSubscriptionConfirmation(ctx context.Context, monitorName, monitorURL string, config map[string]any) error {
	url, err := webhookURL(config, "discord")
	if err != nil {
		return err
	}
	content := fmt.Sprintf("✅ **This channel** is now subscribed to Pingr alerts for **%s** (%s)\nYou'll be notified here when the monitor goes down or recovers.",
		monitorName, monitorURL,
	)
	return post(ctx, n.client, url, map[string]string{"content": content})
}

func (n *discordNotifier) Send(ctx context.Context, event domain.AlertEvent, config map[string]any) error {
	url, err := webhookURL(config, "discord")
	if err != nil {
		return err
	}
	return post(ctx, n.client, url, discordEmbed(fromAlertEvent(event)))
}

func (n *discordNotifier) SendIncidentUpdate(ctx context.Context, incident domain.Incident, update domain.IncidentUpdate, config map[string]any) error {
	url, err := webhookURL(config, "discord")
	if err != nil {
		return err
	}
	return post(ctx, n.client, url, discordEmbed(fromIncidentUpdate(incident, update)))
}

// discordEmbed renders a notification into a Discord embed payload.
func discordEmbed(n notification) map[string]any {
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
	return map[string]any{
		"embeds": []embed{{
			Title:       fmt.Sprintf("%s [%s] %s", n.Emoji, n.TypeTag, n.Name),
			Description: n.Message,
			Color:       n.Color,
			Fields: []field{
				{Name: "Affected", Value: n.Affected, Inline: true},
				{Name: "Time", Value: n.Time, Inline: true},
			},
		}},
	}
}

// ── shared ────────────────────────────────────────────────────────────────────

func webhookURL(config map[string]any, platform string) (string, error) {
	url, ok := config["webhook_url"].(string)
	if !ok || url == "" {
		return "", fmt.Errorf("%s notifier: missing webhook_url in config", platform)
	}
	return url, nil
}

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
