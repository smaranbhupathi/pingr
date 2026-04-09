package email

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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
	html := fmt.Sprintf(`<p>Verify your email: <a href="%s">%s</a></p>`, link, link)
	return c.send(ctx, toEmail, "Verify your email", html)
}

func (c *resendClient) SendPasswordReset(ctx context.Context, toEmail, token string) error {
	link := fmt.Sprintf("%s/reset-password?token=%s", c.appBaseURL, token)
	html := fmt.Sprintf(`<p>Reset your password: <a href="%s">%s</a></p>`, link, link)
	return c.send(ctx, toEmail, "Reset your password", html)
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

func (n *emailNotifier) Send(ctx context.Context, event domain.AlertEvent, config map[string]any) error {
	toEmail, ok := config["email"].(string)
	if !ok || toEmail == "" {
		return fmt.Errorf("email notifier: missing email in config")
	}

	var subject, html string
	switch event.Type {
	case domain.AlertEventDown:
		subject = fmt.Sprintf("🔴 %s is DOWN", event.Monitor.Name)
		html = fmt.Sprintf(`<h2>%s is DOWN</h2><p>URL: %s</p><p>Detected at: %s</p>`,
			event.Monitor.Name, event.Monitor.URL, event.Incident.StartedAt.Format("2006-01-02 15:04:05 UTC"))
	case domain.AlertEventRecovery:
		subject = fmt.Sprintf("🟢 %s is back UP", event.Monitor.Name)
		html = fmt.Sprintf(`<h2>%s is back UP</h2><p>URL: %s</p><p>Resolved at: %s</p>`,
			event.Monitor.Name, event.Monitor.URL, event.Incident.ResolvedAt.Format("2006-01-02 15:04:05 UTC"))
	}

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
		return fmt.Errorf("resend API error: status %d", resp.StatusCode)
	}
	return nil
}
