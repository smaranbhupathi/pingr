package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/smaranbhupathi/pingr/internal/adapters/inbound/http/handler"
	"github.com/smaranbhupathi/pingr/internal/adapters/inbound/http/middleware"
)

func NewRouter(
	authH *handler.AuthHandler,
	monitorH *handler.MonitorHandler,
	userH *handler.UserHandler,
	jwtSecret string,
) http.Handler {
	r := chi.NewRouter()

	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.RealIP)
	r.Use(corsMiddleware)

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

		// Alert channels
		r.Route("/alert-channels", func(r chi.Router) {
			r.Get("/", userH.ListAlertChannels)
			r.Post("/", userH.CreateAlertChannel)
			r.Delete("/{id}", userH.DeleteAlertChannel)
		})

		// Monitors
		r.Route("/monitors", func(r chi.Router) {
			r.Get("/", monitorH.List)
			r.Post("/", monitorH.Create)
			r.Get("/{id}", monitorH.Get)
			r.Patch("/{id}", monitorH.Update)
			r.Delete("/{id}", monitorH.Delete)
			r.Get("/{id}/graph", monitorH.ResponseTimeGraph)
			r.Post("/{id}/subscribe", userH.SubscribeMonitorToChannel)
		})
	})

	return r
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
