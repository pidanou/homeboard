package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/pidanou/family-board/internal/service"
)

type InviteHandler struct {
	invites   *service.InviteService
	families  *service.HouseholdService
	jwtSecret string
}

func NewInviteHandler(invites *service.InviteService, families *service.HouseholdService, jwtSecret string) *InviteHandler {
	return &InviteHandler{invites: invites, families: families, jwtSecret: jwtSecret}
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
