package middleware

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
)

const RequestIDKey contextKey = "request_id"

// RequestLogger logs every HTTP request with method, path, status, latency and request ID.
// For 4xx/5xx responses it also logs the error message from the response body.
// In production these are JSON lines consumable by any log aggregator.
func RequestLogger(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Attach a short request ID so all logs within a request share the same ID
			reqID := uuid.New().String()[:8]
			ctx := context.WithValue(r.Context(), RequestIDKey, reqID)
			w.Header().Set("X-Request-Id", reqID)

			ww := &responseWriter{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(ww, r.WithContext(ctx))

			latency := time.Since(start)
			level := slog.LevelInfo
			if ww.status >= 500 {
				level = slog.LevelError
			} else if ww.status >= 400 {
				level = slog.LevelWarn
			}

			attrs := []slog.Attr{
				slog.String("request_id", reqID),
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", ww.status),
				slog.Int64("latency_ms", latency.Milliseconds()),
				// slog.String("ip", r.RemoteAddr),
				// slog.String("user_agent", r.UserAgent()),
			}

			// For error responses, extract and log the error message from the body
			if ww.status >= 400 && len(ww.errBody) > 0 {
				var payload map[string]string
				if json.Unmarshal(ww.errBody, &payload) == nil {
					if msg := payload["error"]; msg != "" {
						attrs = append(attrs, slog.String("error", msg))
					}
				}
			}

			log.LogAttrs(ctx, level, "request", attrs...)
		})
	}
}

// RequestIDFromContext extracts the request ID — use in handlers for error logs.
func RequestIDFromContext(ctx context.Context) string {
	id, _ := ctx.Value(RequestIDKey).(string)
	return id
}

// responseWriter wraps http.ResponseWriter to capture status code and error body.
type responseWriter struct {
	http.ResponseWriter
	status  int
	errBody []byte // captures up to 512 bytes of the body on error responses
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if rw.status >= 400 && len(rw.errBody) < 512 {
		rw.errBody = append(rw.errBody, b...)
	}
	return rw.ResponseWriter.Write(b)
}
