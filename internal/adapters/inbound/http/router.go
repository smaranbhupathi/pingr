package http

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/smaranbhupathi/pingr/internal/adapters/inbound/http/handler"
	"github.com/smaranbhupathi/pingr/internal/adapters/inbound/http/middleware"
	"github.com/smaranbhupathi/pingr/internal/adapters/inbound/http/ratelimit"
)

func NewRouter(
	authH *handler.AuthHandler,
	monitorH *handler.MonitorHandler,
	userH *handler.UserHandler,
	incidentH *handler.IncidentHandler,
	componentH *handler.ComponentHandler,
	jwtSecret string,
	allowedOrigin string,
	rlStore ratelimit.Store,
	log *slog.Logger,
) http.Handler {
	r := chi.NewRouter()

	// RealIP must run first — it sets r.RemoteAddr to the real client IP
	// by unwrapping X-Forwarded-For. The rate limiter reads r.RemoteAddr,
	// so this order matters.
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Recoverer)
	r.Use(middleware.RequestLogger(log))
	r.Use(corsMiddleware(allowedOrigin))
	r.Use(ratelimit.Middleware(rlStore))

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Public auth routes
	r.Route("/api/v1/auth", func(r chi.Router) {
		r.Post("/register", authH.Register)
		r.Post("/login", authH.Login)
		r.Post("/refresh", authH.Refresh)
		r.Get("/verify-email", authH.VerifyEmail)
		r.Post("/forgot-password", authH.ForgotPassword)
		r.Post("/reset-password", authH.ResetPassword)
	})

	// Public status page (no auth)
	r.Get("/api/v1/status/{username}", monitorH.StatusPage)

	// Protected routes
	r.Route("/api/v1", func(r chi.Router) {
		r.Use(middleware.Authenticate(jwtSecret))

		// Current user
		r.Get("/me", userH.Me)
		r.Post("/me/avatar-upload-url", userH.AvatarUploadURL)
		r.Patch("/me/avatar", userH.UpdateAvatar)
		r.Put("/me/slug", userH.SetSlug)

		// Alert channels
		r.Route("/alert-channels", func(r chi.Router) {
			r.Get("/", userH.ListAlertChannels)
			r.Post("/", userH.CreateAlertChannel)
			r.Post("/import", userH.ImportAlertChannels)
			r.Get("/{id}", userH.GetAlertChannel)
			r.Patch("/{id}", userH.UpdateAlertChannel)
			r.Delete("/{id}", userH.DeleteAlertChannel)
		})

		// Incidents
		r.Route("/incidents", func(r chi.Router) {
			r.Get("/", incidentH.List)
			r.Post("/", incidentH.Create)
			r.Get("/{id}", incidentH.Get)
			r.Post("/{id}/updates", incidentH.PostUpdate)
		})

		// Monitors
		r.Route("/monitors", func(r chi.Router) {
			r.Get("/", monitorH.List)
			r.Post("/", monitorH.Create)
			r.Get("/{id}", monitorH.Get)
			r.Patch("/{id}", monitorH.Update)
			r.Patch("/{id}/meta", userH.UpdateMonitorMeta)
			r.Delete("/{id}", monitorH.Delete)
			r.Get("/{id}/graph", monitorH.ResponseTimeGraph)
			r.Get("/{id}/subscriptions", userH.ListMonitorSubscriptions)
			r.Post("/{id}/subscribe", userH.SubscribeMonitorToChannel)
			r.Delete("/{id}/subscriptions/{channelId}", userH.UnsubscribeMonitorFromChannel)
		})

		// Components
		r.Route("/components", func(r chi.Router) {
			r.Get("/", componentH.List)
			r.Post("/", componentH.Create)
			r.Patch("/{id}", componentH.Update)
			r.Delete("/{id}", componentH.Delete)
		})
	})

	return r
}

// corsMiddleware returns a middleware that sets CORS headers.
// allowedOrigin should be "*" in dev and your exact frontend URL in prod
// (e.g. "https://pingr.yourdomain.com"). Locking to a specific origin in prod
// prevents other websites from making authenticated requests on behalf of your users.
//
// In addition to the configured allowedOrigin, any subdomain of getpingr.com is
// also allowed — status pages are served from username.getpingr.com and the
// React app on those subdomains needs to reach the API.
func corsMiddleware(allowedOrigin string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			allowed := allowedOrigin
			if allowedOrigin == "*" || origin == allowedOrigin ||
				strings.HasSuffix(origin, ".getpingr.com") {
				allowed = origin
			}
			w.Header().Set("Access-Control-Allow-Origin", allowed)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
