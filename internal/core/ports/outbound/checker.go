package outbound

import (
	"context"

	"github.com/smaranbhupathi/pingr/internal/core/domain"
)

// CheckResult is returned by a Checker after performing a monitor check.
type CheckResult struct {
	IsUp           bool
	StatusCode     *int
	ResponseTimeMs int64
	ErrorMessage   string
}

// Checker performs the actual connectivity check for a given monitor.
// Roll-out 1: HTTPChecker
// Roll-out 2: TCPChecker, DNSChecker
type Checker interface {
	// Type returns the monitor type this checker handles.
	Type() domain.MonitorType

	// Check performs a single check against the monitor's target.
	Check(ctx context.Context, monitor domain.Monitor) (CheckResult, error)
}
