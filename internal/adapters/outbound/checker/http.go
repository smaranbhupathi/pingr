package checker

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/smaranbhupathi/pingr/internal/core/domain"
	"github.com/smaranbhupathi/pingr/internal/core/ports/outbound"
)

type httpChecker struct {
	client *http.Client
}

// NewHTTPChecker returns an outbound.Checker for HTTP/HTTPS monitors.
func NewHTTPChecker() outbound.Checker {
	return &httpChecker{
		client: &http.Client{
			Timeout: 30 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 5 {
					return fmt.Errorf("too many redirects")
				}
				return nil
			},
		},
	}
}

func (c *httpChecker) Type() domain.MonitorType {
	return domain.MonitorTypeHTTP
}

func (c *httpChecker) Check(ctx context.Context, monitor domain.Monitor) (outbound.CheckResult, error) {
	timeout := time.Duration(monitor.TimeoutSeconds) * time.Second
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, monitor.URL, nil)
	if err != nil {
		return outbound.CheckResult{IsUp: false, ErrorMessage: err.Error()}, nil
	}
	req.Header.Set("User-Agent", "Upmon-Monitor/1.0")

	start := time.Now()
	resp, err := c.client.Do(req)
	responseTimeMs := time.Since(start).Milliseconds()

	if err != nil {
		return outbound.CheckResult{
			IsUp:           false,
			ResponseTimeMs: responseTimeMs,
			ErrorMessage:   err.Error(),
		}, nil
	}
	defer resp.Body.Close()

	statusCode := resp.StatusCode
	isUp := statusCode >= 200 && statusCode < 400

	result := outbound.CheckResult{
		IsUp:           isUp,
		StatusCode:     &statusCode,
		ResponseTimeMs: responseTimeMs,
	}
	if !isUp {
		result.ErrorMessage = fmt.Sprintf("HTTP %d", statusCode)
	}

	return result, nil
}
