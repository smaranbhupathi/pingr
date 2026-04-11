package email

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/smaranbhupathi/pingr/internal/core/domain"
	"github.com/smaranbhupathi/pingr/internal/core/ports/outbound"
)

const resendAPI = "https://api.resend.com/emails"

type resendClient struct {
	apiKey     string
	fromEmail  string
	appBaseURL string
	httpClient *http.Client
}

type resendPayload struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	HTML    string   `json:"html"`
}

// NewEmailSender returns an outbound.EmailSender backed by Resend.com.
func NewEmailSender(apiKey, fromEmail, appBaseURL string) outbound.EmailSender {
	return &resendClient{
		apiKey:     apiKey,
		fromEmail:  fromEmail,
		appBaseURL: appBaseURL,
		httpClient: &http.Client{},
	}
}

// NewNotifier returns an outbound.Notifier for email alert channels.
func NewNotifier(apiKey, fromEmail string) outbound.Notifier {
	return &emailNotifier{apiKey: apiKey, fromEmail: fromEmail, httpClient: &http.Client{}}
}

func (c *resendClient) SendVerification(ctx context.Context, toEmail, token string) error {
	link := fmt.Sprintf("%s/verify-email?token=%s", c.appBaseURL, token)
	html := fmt.Sprintf(`<!DOCTYPE html>
<html><body style="font-family:sans-serif;background:#f9fafb;margin:0;padding:40px 0;">
<div style="max-width:480px;margin:0 auto;background:white;border-radius:12px;padding:40px;border:1px solid #e5e7eb;">
  <h1 style="color:#4f46e5;font-size:22px;margin:0 0 4px;">Pingr</h1>
  <p style="color:#6b7280;font-size:13px;margin:0 0 28px;">Uptime monitoring, simplified.</p>
  <h2 style="color:#111827;font-size:18px;margin:0 0 12px;">Verify your email address</h2>
  <p style="color:#374151;margin:0 0 24px;line-height:1.6;">
    Thanks for signing up! Click the button below to confirm your email address and activate your account.
  </p>
  <a href="%s" style="display:inline-block;background:#4f46e5;color:white;padding:12px 28px;border-radius:8px;text-decoration:none;font-weight:600;font-size:15px;">
    Verify email →
  </a>
  <p style="color:#9ca3af;font-size:12px;margin:28px 0 8px;">This link expires in 24 hours. If you didn't create a Pingr account, you can ignore this email.</p>
  <p style="color:#d1d5db;font-size:11px;word-break:break-all;">Or paste this link in your browser:<br>%s</p>
</div>
</body></html>`, link, link)
	return c.send(ctx, toEmail, "Verify your Pingr email", html)
}

func (c *resendClient) SendPasswordReset(ctx context.Context, toEmail, token string) error {
	link := fmt.Sprintf("%s/reset-password?token=%s", c.appBaseURL, token)
	html := fmt.Sprintf(`<!DOCTYPE html>
<html><body style="font-family:sans-serif;background:#f9fafb;margin:0;padding:40px 0;">
<div style="max-width:480px;margin:0 auto;background:white;border-radius:12px;padding:40px;border:1px solid #e5e7eb;">
  <h1 style="color:#4f46e5;font-size:22px;margin:0 0 4px;">Pingr</h1>
  <p style="color:#6b7280;font-size:13px;margin:0 0 28px;">Uptime monitoring, simplified.</p>
  <h2 style="color:#111827;font-size:18px;margin:0 0 12px;">Reset your password</h2>
  <p style="color:#374151;margin:0 0 24px;line-height:1.6;">
    We received a request to reset your Pingr password. Click the button below to choose a new password.
  </p>
  <a href="%s" style="display:inline-block;background:#4f46e5;color:white;padding:12px 28px;border-radius:8px;text-decoration:none;font-weight:600;font-size:15px;">
    Reset password →
  </a>
  <p style="color:#9ca3af;font-size:12px;margin:28px 0 8px;">This link expires in 1 hour. If you didn't request a password reset, you can safely ignore this email.</p>
  <p style="color:#d1d5db;font-size:11px;word-break:break-all;">Or paste this link in your browser:<br>%s</p>
</div>
</body></html>`, link, link)
	return c.send(ctx, toEmail, "Reset your Pingr password", html)
}

func (c *resendClient) SendSubscriptionConfirmation(ctx context.Context, toEmail, monitorName, monitorURL string) error {
	html := fmt.Sprintf(`<!DOCTYPE html>
<html><body style="font-family:sans-serif;background:#f9fafb;margin:0;padding:40px 0;">
<div style="max-width:480px;margin:0 auto;background:white;border-radius:12px;padding:40px;border:1px solid #e5e7eb;">
  <h1 style="color:#4f46e5;font-size:22px;margin:0 0 4px;">Pingr</h1>
  <p style="color:#6b7280;font-size:13px;margin:0 0 28px;">Uptime monitoring, simplified.</p>
  <h2 style="color:#111827;font-size:18px;margin:0 0 12px;">You're now subscribed to alerts</h2>
  <p style="color:#374151;margin:0 0 20px;line-height:1.6;">
    This email address will now receive alerts for:
  </p>
  <div style="background:#f5f3ff;border:1px solid #e0e7ff;border-radius:8px;padding:16px 20px;margin:0 0 24px;">
    <p style="margin:0;font-weight:600;color:#1e1b4b;">%s</p>
    <p style="margin:4px 0 0;color:#6b7280;font-size:13px;">%s</p>
  </div>
  <p style="color:#374151;margin:0 0 8px;line-height:1.6;">You'll get an email when:</p>
  <ul style="color:#374151;margin:0 0 24px;padding-left:20px;line-height:1.8;">
    <li>🔴 The monitor goes <strong>DOWN</strong></li>
    <li>🟢 The monitor <strong>recovers</strong></li>
  </ul>
  <p style="color:#9ca3af;font-size:12px;margin:0;">To stop receiving alerts, remove this email from the monitor's alert channels in your Pingr dashboard.</p>
</div>
</body></html>`, monitorName, monitorURL)
	return c.send(ctx, toEmail, fmt.Sprintf("Subscribed to alerts for %s", monitorName), html)
}

func (c *resendClient) send(ctx context.Context, to, subject, html string) error {
	return sendViaResend(ctx, c.httpClient, c.apiKey, c.fromEmail, to, subject, html)
}

// --- Alert Notifier ---

type emailNotifier struct {
	apiKey     string
	fromEmail  string
	httpClient *http.Client
}

func (n *emailNotifier) Type() domain.AlertChannelType {
	return domain.AlertChannelEmail
}

// SendSubscriptionConfirmation is a no-op for email — the EmailSender already
// sends a dedicated confirmation email via SendSubscriptionConfirmation.
func (n *emailNotifier) SendSubscriptionConfirmation(_ context.Context, _, _ string, _ map[string]any) error {
	return nil
}

func (n *emailNotifier) Send(ctx context.Context, event domain.AlertEvent, config map[string]any) error {
	toEmail, ok := config["email"].(string)
	if !ok || toEmail == "" {
		return fmt.Errorf("email notifier: missing email in config")
	}
	return n.sendNotification(ctx, toEmail, emailNotifFromAlertEvent(event))
}

func (n *emailNotifier) SendIncidentUpdate(ctx context.Context, incident domain.Incident, update domain.IncidentUpdate, config map[string]any) error {
	toEmail, ok := config["email"].(string)
	if !ok || toEmail == "" {
		return fmt.Errorf("email notifier: missing email in config")
	}
	return n.sendNotification(ctx, toEmail, emailNotifFromIncidentUpdate(incident, update))
}

// emailNotif holds the normalised fields for the shared HTML template.
type emailNotif struct {
	Emoji    string
	TypeTag  string
	Name     string
	Message  string
	Affected string
	Time     string
	HexColor string // border + badge colour
}

func emailNotifFromAlertEvent(event domain.AlertEvent) emailNotif {
	switch event.Type {
	case domain.AlertEventDown:
		return emailNotif{
			Emoji:    "🔴",
			TypeTag:  "DOWN",
			Name:     event.Monitor.Name,
			Message:  fmt.Sprintf("%s is unreachable.", event.Monitor.Name),
			Affected: fmt.Sprintf("%s — %s", event.Monitor.Name, event.Monitor.URL),
			Time:     event.OutageEvent.StartedAt.UTC().Format("2006-01-02 15:04:05 UTC"),
			HexColor: "#dc2626",
		}
	case domain.AlertEventRecovery:
		return emailNotif{
			Emoji:    "🟢",
			TypeTag:  "RECOVERED",
			Name:     event.Monitor.Name,
			Message:  fmt.Sprintf("%s has recovered and is back online.", event.Monitor.Name),
			Affected: fmt.Sprintf("%s — %s", event.Monitor.Name, event.Monitor.URL),
			Time:     event.OutageEvent.ResolvedAt.UTC().Format("2006-01-02 15:04:05 UTC"),
			HexColor: "#16a34a",
		}
	default:
		return emailNotif{}
	}
}

func emailNotifFromIncidentUpdate(incident domain.Incident, update domain.IncidentUpdate) emailNotif {
	affected := "—"
	if len(incident.Monitors) > 0 {
		parts := make([]string, 0, len(incident.Monitors))
		for _, m := range incident.Monitors {
			parts = append(parts, m.Name)
		}
		affected = strings.Join(parts, ", ")
	}

	typeTag, hexColor := incidentEmailMeta(update.Status)
	emoji := incidentEmailEmoji(update.Status)
	return emailNotif{
		Emoji:    emoji,
		TypeTag:  typeTag,
		Name:     incident.Name,
		Message:  update.Message,
		Affected: affected,
		Time:     update.CreatedAt.UTC().Format("2006-01-02 15:04:05 UTC"),
		HexColor: hexColor,
	}
}

func incidentEmailMeta(s domain.IncidentStatus) (typeTag, hexColor string) {
	switch s {
	case domain.IncidentStatusInvestigating:
		return "INVESTIGATING", "#dc2626"
	case domain.IncidentStatusIdentified:
		return "IDENTIFIED", "#f97316"
	case domain.IncidentStatusMonitoring:
		return "MONITORING", "#eab308"
	case domain.IncidentStatusResolved:
		return "RESOLVED", "#16a34a"
	default:
		return strings.ToUpper(string(s)), "#6b7280"
	}
}

func incidentEmailEmoji(s domain.IncidentStatus) string {
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

// sendNotification renders the shared HTML template and dispatches via Resend.
func (n *emailNotifier) sendNotification(ctx context.Context, toEmail string, notif emailNotif) error {
	subject := fmt.Sprintf("%s [%s] %s", notif.Emoji, notif.TypeTag, notif.Name)
	html := fmt.Sprintf(`<!DOCTYPE html>
<html><body style="font-family:sans-serif;background:#f9fafb;margin:0;padding:40px 0;">
<div style="max-width:500px;margin:0 auto;background:white;border-radius:12px;padding:40px;border:1px solid #e5e7eb;">

  <h1 style="color:#4f46e5;font-size:22px;margin:0 0 2px;">Pingr</h1>
  <p style="color:#9ca3af;font-size:12px;margin:0 0 28px;">Uptime monitoring, simplified.</p>

  <div style="display:flex;align-items:center;gap:10px;margin:0 0 16px;">
    <span style="display:inline-block;padding:3px 10px;border-radius:20px;font-size:11px;font-weight:700;letter-spacing:.5px;background:%s;color:white;">%s</span>
    <h2 style="font-size:17px;color:#111827;margin:0;font-weight:600;">%s</h2>
  </div>

  <div style="background:#f8fafc;border-left:3px solid %s;border-radius:4px;padding:14px 18px;margin:0 0 24px;">
    <p style="margin:0;color:#374151;font-size:14px;line-height:1.7;">%s</p>
  </div>

  <table style="width:100%%;border-collapse:collapse;font-size:13px;">
    <tr style="border-top:1px solid #f3f4f6;">
      <td style="padding:10px 0 10px;color:#6b7280;font-weight:600;width:100px;">Affected</td>
      <td style="padding:10px 0 10px;color:#374151;">%s</td>
    </tr>
    <tr style="border-top:1px solid #f3f4f6;">
      <td style="padding:10px 0;color:#6b7280;font-weight:600;">Time</td>
      <td style="padding:10px 0;color:#374151;">%s</td>
    </tr>
  </table>

</div>
</body></html>`,
		notif.HexColor, notif.TypeTag,
		notif.Name,
		notif.HexColor,
		notif.Message,
		notif.Affected,
		notif.Time,
	)

	return sendViaResend(ctx, n.httpClient, n.apiKey, n.fromEmail, toEmail, subject, html)
}

func sendViaResend(ctx context.Context, client *http.Client, apiKey, from, to, subject, html string) error {
	payload := resendPayload{From: from, To: []string{to}, Subject: subject, HTML: html}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, resendAPI, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("resend request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("resend error %d: %s", resp.StatusCode, string(respBody))
	}
	return nil
}
