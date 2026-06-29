package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/pidanou/homeboard/internal/service"
)

type InviteHandler struct {
	invites   *service.InviteService
	families  *service.HouseholdService
	auth      *service.AuthService
	jwtSecret string
}

func NewInviteHandler(invites *service.InviteService, families *service.HouseholdService, auth *service.AuthService, jwtSecret string) *InviteHandler {
	return &InviteHandler{invites: invites, families: families, auth: auth, jwtSecret: jwtSecret}
}

func (h *InviteHandler) Routes() http.Handler {
	r := chi.NewRouter()
	r.Get("/", h.list)
	r.Post("/", h.create)
	r.Delete("/{token}", h.delete)
	return r
}

func (h *InviteHandler) PublicRoutes() http.Handler {
	r := chi.NewRouter()
	r.Get("/{token}", h.get)
	r.Post("/{token}/register", h.registerAndAccept)
	r.Group(func(r chi.Router) {
		r.Use(AuthMiddleware(h.jwtSecret))
		r.Post("/{token}/accept", h.accept)
	})
	return r
}

func (h *InviteHandler) delete(w http.ResponseWriter, r *http.Request) {
	familyID := chi.URLParam(r, "familyID")
	if err := requireAdmin(r, familyID, h.families); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	token := chi.URLParam(r, "token")
	if err := h.invites.Delete(r.Context(), token); err != nil {
		http.Error(w, "failed to delete invite", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *InviteHandler) list(w http.ResponseWriter, r *http.Request) {
	familyID := chi.URLParam(r, "familyID")
	if err := requireAdmin(r, familyID, h.families); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	invites, err := h.invites.ListForFamily(r.Context(), familyID)
	if err != nil {
		http.Error(w, "failed to list invites", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(invites)
}

func (h *InviteHandler) create(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(ContextKeyUserID).(string)
	familyID := chi.URLParam(r, "familyID")
	if err := requireAdmin(r, familyID, h.families); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	invite, err := h.invites.Create(r.Context(), familyID, userID)
	if err != nil {
		http.Error(w, "failed to create invite", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(invite)
}

func (h *InviteHandler) get(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")

	invite, err := h.invites.GetByToken(r.Context(), token)
	if err != nil {
		http.Error(w, "invite not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(invite)
}

func (h *InviteHandler) registerAndAccept(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")

	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Name     string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	// Fully validate invite (exists, not used, not expired) before creating the account
	if err := h.invites.Validate(r.Context(), token); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := h.auth.CreateUser(r.Context(), body.Email, body.Password, body.Name)
	if err != nil {
		http.Error(w, "registration failed", http.StatusInternalServerError)
		return
	}

	result, err := h.invites.Accept(r.Context(), token, user.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	jwt, err := h.auth.IssueToken(user.ID)
	if err != nil {
		http.Error(w, "failed to issue token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]any{
		"token":     jwt,
		"family_id": result.FamilyID,
		"unlinked_virtual_members": result.UnlinkedVirtualMembers,
	})
}

func (h *InviteHandler) accept(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(ContextKeyUserID).(string)
	if !ok || userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	token := chi.URLParam(r, "token")

	result, err := h.invites.Accept(r.Context(), token, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
