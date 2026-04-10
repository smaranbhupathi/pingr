package ratelimit

import "time"

// RouteLimit defines the rate limit for a specific route.
type RouteLimit struct {
	Limit  int           // maximum number of requests allowed
	Window time.Duration // within this rolling time window
}

// Routes maps exact URL paths to their rate limits.
//
// ── HOW TO ADD A NEW RULE ─────────────────────────────────────────────────────
// Add a line here. The key is the exact URL path. That's it.
// No other file needs to change.
//
//   "/api/v1/some/new/endpoint": {Limit: 20, Window: time.Minute},
//
// ── HOW TO CHANGE AN EXISTING LIMIT ──────────────────────────────────────────
// Edit the Limit or Window value. Redeploy. Done.
//
// ── FALLBACK ──────────────────────────────────────────────────────────────────
// Any route not listed here falls back to the "default" entry.
// ─────────────────────────────────────────────────────────────────────────────

var Routes = map[string]RouteLimit{
	// Auth routes — tight limits to block brute force and spam.

	// Login: 10 attempts per minute per IP.
	// An attacker trying common passwords would be blocked after 10 tries.
	"/api/v1/auth/login": {Limit: 10, Window: time.Minute},

	// Register: 5 accounts per hour per IP.
	// Slows down bot-driven account creation.
	"/api/v1/auth/register": {Limit: 5, Window: time.Hour},

	// Forgot password: 3 requests per hour per IP.
	// Each request sends an email — this prevents email flooding a victim.
	"/api/v1/auth/forgot-password": {Limit: 3, Window: time.Hour},

	// Reset password: 5 attempts per hour per IP.
	// Tokens are already single-use and expiring, but belt-and-suspenders.
	"/api/v1/auth/reset-password": {Limit: 5, Window: time.Hour},

	// Default: 100 requests per minute for all other routes.
	// Covers monitor CRUD, alert channels, profile — normal API usage.
	"default": {Limit: 100, Window: time.Minute},
}
