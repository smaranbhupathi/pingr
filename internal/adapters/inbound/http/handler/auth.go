package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/smaranbhupathi/pingr/internal/adapters/inbound/http/middleware"
	"github.com/smaranbhupathi/pingr/internal/core/ports/inbound"
	"github.com/smaranbhupathi/pingr/internal/core/services"
)

type AuthHandler struct {
	auth inbound.AuthService
	log  *slog.Logger
}

func NewAuthHandler(auth inbound.AuthService, log *slog.Logger) *AuthHandler {
	return &AuthHandler{auth: auth, log: log}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var input inbound.RegisterInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if errs := input.Validate(); len(errs) > 0 {
		JSON(w, http.StatusUnprocessableEntity, map[string]any{"errors": errs})
		return
	}

	if err := h.auth.Register(r.Context(), input); err != nil {
		switch {
		case errors.Is(err, services.ErrEmailTaken):
			Error(w, http.StatusConflict, "email already registered")
		case errors.Is(err, services.ErrUsernameTaken):
			Error(w, http.StatusConflict, "username already taken")
		default:
			h.log.ErrorContext(r.Context(), "register failed",
				"request_id", middleware.RequestIDFromContext(r.Context()),
				"email", input.Email,
				"error", err,
			)
			Error(w, http.StatusInternalServerError, "registration failed")
		}
		return
	}

	h.log.InfoContext(r.Context(), "user registered",
		"request_id", middleware.RequestIDFromContext(r.Context()),
		"email", input.Email,
		"username", input.Username,
	)
	JSON(w, http.StatusCreated, map[string]string{"message": "check your email to verify your account"})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var input inbound.LoginInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if errs := input.Validate(); len(errs) > 0 {
		JSON(w, http.StatusUnprocessableEntity, map[string]any{"errors": errs})
		return
	}

	tokens, err := h.auth.Login(r.Context(), input)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidCreds):
			Error(w, http.StatusUnauthorized, "invalid email or password")
		case errors.Is(err, services.ErrNotVerified):
			Error(w, http.StatusForbidden, "please verify your email first")
		default:
			h.log.ErrorContext(r.Context(), "login failed",
				"request_id", middleware.RequestIDFromContext(r.Context()),
				"email", input.Email,
				"error", err,
			)
			Error(w, http.StatusInternalServerError, "login failed")
		}
		return
	}

	h.log.InfoContext(r.Context(), "user logged in",
		"request_id", middleware.RequestIDFromContext(r.Context()),
		"email", input.Email,
	)
	JSON(w, http.StatusOK, tokens)
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var body struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tokens, err := h.auth.RefreshTokens(r.Context(), body.RefreshToken)
	if err != nil {
		h.log.WarnContext(r.Context(), "token refresh rejected — invalid or expired",
			"request_id", middleware.RequestIDFromContext(r.Context()),
		)
		Error(w, http.StatusUnauthorized, "invalid or expired refresh token")
		return
	}

	h.log.InfoContext(r.Context(), "tokens refreshed",
		"request_id", middleware.RequestIDFromContext(r.Context()),
	)
	JSON(w, http.StatusOK, tokens)
}

func (h *AuthHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		Error(w, http.StatusBadRequest, "missing token")
		return
	}

	if err := h.auth.VerifyEmail(r.Context(), token); err != nil {
		Error(w, http.StatusBadRequest, "invalid or expired token")
		return
	}

	h.log.InfoContext(r.Context(), "email verified",
		"request_id", middleware.RequestIDFromContext(r.Context()),
	)
	JSON(w, http.StatusOK, map[string]string{"message": "email verified successfully"})
}

func (h *AuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	h.auth.ForgotPassword(r.Context(), body.Email)
	// Log without the email — we intentionally don't reveal whether the address exists
	h.log.InfoContext(r.Context(), "forgot password requested",
		"request_id", middleware.RequestIDFromContext(r.Context()),
	)
	JSON(w, http.StatusOK, map[string]string{"message": "if that email exists, a reset link has been sent"})
}

func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Token       string `json:"token"`
		NewPassword string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.auth.ResetPassword(r.Context(), body.Token, body.NewPassword); err != nil {
		h.log.WarnContext(r.Context(), "password reset failed — invalid or expired token",
			"request_id", middleware.RequestIDFromContext(r.Context()),
		)
		Error(w, http.StatusBadRequest, "invalid or expired token")
		return
	}

	h.log.InfoContext(r.Context(), "password reset successful",
		"request_id", middleware.RequestIDFromContext(r.Context()),
	)
	JSON(w, http.StatusOK, map[string]string{"message": "password reset successfully"})
}
