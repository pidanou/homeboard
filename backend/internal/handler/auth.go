package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/pidanou/homeboard/internal/service"
)

type AuthHandler struct {
	auth *service.AuthService
}

func NewAuthHandler(auth *service.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

func (h *AuthHandler) Routes() http.Handler {
	r := chi.NewRouter()
	r.Post("/register", h.register)
	r.Post("/login", h.login)
	r.Post("/forgot-password", h.forgotPassword)
	r.Post("/reset-password", h.resetPassword)
	return r
}

func (h *AuthHandler) register(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Name     string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request")
		return
	}

	user, err := h.auth.Register(r.Context(), body.Email, body.Password, body.Name)
	if err != nil {
		status := http.StatusInternalServerError
		if err == service.ErrRegistrationClosed || err == service.ErrPasswordLoginDisabled {
			status = http.StatusForbidden
		}
		writeError(w, status, codeFor(err, "registration_failed"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func (h *AuthHandler) login(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request")
		return
	}

	token, err := h.auth.Login(r.Context(), body.Email, body.Password)
	if err != nil {
		if err == service.ErrPasswordLoginDisabled {
			writeError(w, http.StatusForbidden, "password_login_disabled")
			return
		}
		writeError(w, http.StatusUnauthorized, "invalid_credentials")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}

func (h *AuthHandler) forgotPassword(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request")
		return
	}

	if err := h.auth.RequestPasswordReset(r.Context(), body.Email); err != nil {
		status := http.StatusInternalServerError
		if err == service.ErrPasswordLoginDisabled {
			status = http.StatusForbidden
		}
		writeError(w, status, codeFor(err, "process_request_failed"))
		return
	}

	// Always a generic success — never reveal whether the email exists.
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "if that email exists, a reset link was sent"})
}

func (h *AuthHandler) resetPassword(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Token    string `json:"token"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request")
		return
	}

	if err := h.auth.ResetPassword(r.Context(), body.Token, body.Password); err != nil {
		status := http.StatusInternalServerError
		if err == service.ErrPasswordLoginDisabled {
			status = http.StatusForbidden
		} else if err == service.ErrInvalidResetToken {
			status = http.StatusBadRequest
		}
		writeError(w, status, codeFor(err, "reset_password_failed"))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
