package email

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

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

	var subject, html string
	switch event.Type {
	case domain.AlertEventDown:
		subject = fmt.Sprintf("🔴 %s is DOWN", event.Monitor.Name)
		html = fmt.Sprintf(`<!DOCTYPE html>
<html><body style="font-family:sans-serif;background:#f9fafb;margin:0;padding:40px 0;">
<div style="max-width:480px;margin:0 auto;background:white;border-radius:12px;padding:40px;border:1px solid #fecaca;">
  <h1 style="color:#4f46e5;font-size:22px;margin:0 0 4px;">Pingr</h1>
  <p style="color:#6b7280;font-size:13px;margin:0 0 28px;">Uptime monitoring, simplified.</p>
  <div style="display:flex;align-items:center;gap:8px;margin:0 0 16px;">
    <span style="font-size:28px;">🔴</span>
    <h2 style="color:#dc2626;font-size:20px;margin:0;">%s is DOWN</h2>
  </div>
  <div style="background:#fef2f2;border:1px solid #fecaca;border-radius:8px;padding:16px 20px;margin:0 0 24px;">
    <p style="margin:0;color:#7f1d1d;font-size:13px;font-weight:600;">MONITOR</p>
    <p style="margin:4px 0 0;font-weight:600;color:#1f2937;">%s</p>
    <p style="margin:4px 0 0;color:#6b7280;font-size:13px;">%s</p>
    <p style="margin:8px 0 0;color:#9ca3af;font-size:12px;">Detected at %s UTC</p>
  </div>
  <p style="color:#374151;margin:0;font-size:14px;line-height:1.6;">We'll notify you again when the monitor recovers.</p>
</div>
</body></html>`,
			event.Monitor.Name,
			event.Monitor.Name,
			event.Monitor.URL,
			event.OutageEvent.StartedAt.UTC().Format("2006-01-02 15:04:05"),
		)
	case domain.AlertEventRecovery:
		subject = fmt.Sprintf("🟢 %s is back UP", event.Monitor.Name)
		html = fmt.Sprintf(`<!DOCTYPE html>
<html><body style="font-family:sans-serif;background:#f9fafb;margin:0;padding:40px 0;">
<div style="max-width:480px;margin:0 auto;background:white;border-radius:12px;padding:40px;border:1px solid #bbf7d0;">
  <h1 style="color:#4f46e5;font-size:22px;margin:0 0 4px;">Pingr</h1>
  <p style="color:#6b7280;font-size:13px;margin:0 0 28px;">Uptime monitoring, simplified.</p>
  <div style="display:flex;align-items:center;gap:8px;margin:0 0 16px;">
    <span style="font-size:28px;">🟢</span>
    <h2 style="color:#16a34a;font-size:20px;margin:0;">%s is back UP</h2>
  </div>
  <div style="background:#f0fdf4;border:1px solid #bbf7d0;border-radius:8px;padding:16px 20px;margin:0 0 24px;">
    <p style="margin:0;color:#14532d;font-size:13px;font-weight:600;">MONITOR</p>
    <p style="margin:4px 0 0;font-weight:600;color:#1f2937;">%s</p>
    <p style="margin:4px 0 0;color:#6b7280;font-size:13px;">%s</p>
    <p style="margin:8px 0 0;color:#9ca3af;font-size:12px;">Recovered at %s UTC</p>
  </div>
  <p style="color:#374151;margin:0;font-size:14px;line-height:1.6;">Your monitor is healthy again. 🎉</p>
</div>
</body></html>`,
			event.Monitor.Name,
			event.Monitor.Name,
			event.Monitor.URL,
			event.OutageEvent.ResolvedAt.UTC().Format("2006-01-02 15:04:05"),
		)
	default:
		return fmt.Errorf("unknown alert event type: %s", event.Type)
	}

	return sendViaResend(ctx, n.httpClient, n.apiKey, n.fromEmail, toEmail, subject, html)
}

func (n *emailNotifier) SendIncidentUpdate(ctx context.Context, incident domain.Incident, update domain.IncidentUpdate, config map[string]any) error {
	toEmail, ok := config["email"].(string)
	if !ok || toEmail == "" {
		return fmt.Errorf("email notifier: missing email in config")
	}

	statusLabel := incidentStatusLabel(update.Status)
	statusColor := incidentHexColor(update.Status)
	monitorNames := incidentMonitorNamesHTML(incident)

	subject := fmt.Sprintf("[%s] %s", statusLabel, incident.Name)
	html := fmt.Sprintf(`<!DOCTYPE html>
<html><body style="font-family:sans-serif;background:#f9fafb;margin:0;padding:40px 0;">
<div style="max-width:520px;margin:0 auto;background:white;border-radius:12px;padding:40px;border:1px solid #e5e7eb;">
  <h1 style="color:#4f46e5;font-size:22px;margin:0 0 4px;">Pingr</h1>
  <p style="color:#6b7280;font-size:13px;margin:0 0 28px;">Incident Update</p>

  <div style="display:flex;align-items:center;gap:10px;margin:0 0 20px;">
    <span style="display:inline-block;padding:4px 12px;border-radius:20px;font-size:12px;font-weight:700;background:%s;color:white;">%s</span>
    <h2 style="font-size:18px;color:#111827;margin:0;">%s</h2>
  </div>

  <div style="background:#f8fafc;border:1px solid #e2e8f0;border-radius:8px;padding:16px 20px;margin:0 0 20px;">
    <p style="margin:0;color:#374151;font-size:14px;line-height:1.7;">%s</p>
  </div>

  <table style="width:100%%;border-collapse:collapse;font-size:13px;color:#6b7280;">
    <tr>
      <td style="padding:6px 0;font-weight:600;color:#374151;width:120px;">Affected</td>
      <td style="padding:6px 0;">%s</td>
    </tr>
    <tr>
      <td style="padding:6px 0;font-weight:600;color:#374151;">Updated at</td>
      <td style="padding:6px 0;">%s UTC</td>
    </tr>
  </table>
</div>
</body></html>`,
		statusColor, statusLabel,
		incident.Name,
		update.Message,
		monitorNames,
		update.CreatedAt.UTC().Format("2006-01-02 15:04:05"),
	)

	return sendViaResend(ctx, n.httpClient, n.apiKey, n.fromEmail, toEmail, subject, html)
}

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

func incidentHexColor(s domain.IncidentStatus) string {
	switch s {
	case domain.IncidentStatusInvestigating:
		return "#dc2626"
	case domain.IncidentStatusIdentified:
		return "#f97316"
	case domain.IncidentStatusMonitoring:
		return "#eab308"
	case domain.IncidentStatusResolved:
		return "#16a34a"
	default:
		return "#6b7280"
	}
}

func incidentMonitorNamesHTML(incident domain.Incident) string {
	if len(incident.Monitors) == 0 {
		return "—"
	}
	result := ""
	for i, m := range incident.Monitors {
		if i > 0 {
			result += ", "
		}
		result += m.Name
	}
	return result
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
