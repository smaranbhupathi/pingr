package ratelimit

import (
	"fmt"
	"net"
	"net/http"
)

// Middleware returns a Chi-compatible middleware that enforces per-route
// rate limits defined in config.go using the provided Store.
//
// Changing the algorithm:  pass a different Store implementation.
// Changing the limits:     edit config.go.
// Switching to Redis:      call NewRedisStore() instead of NewMemoryStore() in main.go.
func Middleware(store Store) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := extractIP(r)
			cfg := routeLimit(r.URL.Path)

			// Key is "ip:path" so each (client, route) pair has its own bucket.
			// A single IP is limited per route independently —
			// hitting /login a lot doesn't affect their /monitors quota.
			key := fmt.Sprintf("%s:%s", ip, r.URL.Path)

			allowed, err := store.Allow(r.Context(), key, cfg.Limit, cfg.Window)
			if err != nil {
				// Store errors are non-fatal — fail open (allow the request).
				// Better to let a request through than block legitimate users
				// because of an internal bookkeeping error.
				next.ServeHTTP(w, r)
				return
			}

			if !allowed {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", cfg.Limit))
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte(`{"error":"too many requests, please slow down"}`))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// routeLimit returns the RouteLimit for path, falling back to "default".
func routeLimit(path string) RouteLimit {
	if cfg, ok := Routes[path]; ok {
		return cfg
	}
	return Routes["default"]
}

// extractIP returns the client's IP address from the request.
// Chi's RealIP middleware (registered before this one in the router)
// already unwraps X-Forwarded-For and X-Real-IP headers set by Fly.io's
// proxy layer, so r.RemoteAddr already holds the real client IP.
func extractIP(r *http.Request) string {
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		// RemoteAddr has no port (shouldn't happen, but safe fallback)
		return r.RemoteAddr
	}
	return ip
}
