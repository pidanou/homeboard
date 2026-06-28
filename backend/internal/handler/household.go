package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/pidanou/homeboard/internal/service"
)

type HouseholdHandler struct {
	families *service.HouseholdService
}

func NewHouseholdHandler(families *service.HouseholdService) *HouseholdHandler {
	return &HouseholdHandler{families: families}
}

func (h *HouseholdHandler) Routes() http.Handler {
	r := chi.NewRouter()
	r.Get("/", h.list)
	r.Post("/", h.create)
	r.Get("/{familyID}", h.get)
	r.Patch("/{familyID}", h.updateName)
	r.Get("/{familyID}/members", h.members)
	r.Post("/{familyID}/members/virtual", h.createVirtual)
	r.Delete("/{familyID}/members/{memberID}", h.removeMember)
	r.Put("/{familyID}/members/{memberID}/role", h.updateRole)
	r.Delete("/{familyID}/members/virtual/{memberID}", h.deleteVirtual)
	r.Post("/{familyID}/members/virtual/{memberID}/link", h.linkVirtual)
	r.Get("/{familyID}/members/virtual/unlinked", h.unlinkedVirtual)
	return r
}

func (h *HouseholdHandler) list(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(ContextKeyUserID).(string)

	families, err := h.families.ListForUser(r.Context(), userID)
	if err != nil {
		http.Error(w, "failed to list families", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(families)
}

func (h *HouseholdHandler) create(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(ContextKeyUserID).(string)

	var body struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	family, err := h.families.Create(r.Context(), body.Name, userID)
	if err != nil {
		http.Error(w, "failed to create family", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(family)
}

func (h *HouseholdHandler) get(w http.ResponseWriter, r *http.Request) {
	familyID := chi.URLParam(r, "familyID")
	if err := requireMember(r, familyID, h.families); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	family, err := h.families.GetByID(r.Context(), familyID)
	if err != nil {
		http.Error(w, "family not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(family)
}

func (h *HouseholdHandler) members(w http.ResponseWriter, r *http.Request) {
	familyID := chi.URLParam(r, "familyID")
	if err := requireMember(r, familyID, h.families); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	members, err := h.families.GetMembers(r.Context(), familyID)
	if err != nil {
		http.Error(w, "failed to get members", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(members)
}

func (h *HouseholdHandler) createVirtual(w http.ResponseWriter, r *http.Request) {
	familyID := chi.URLParam(r, "familyID")
	callerID, ok := r.Context().Value(ContextKeyUserID).(string)
	if !ok || callerID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	var body struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name == "" {
		http.Error(w, "name required", http.StatusBadRequest)
		return
	}
	m, err := h.families.CreateVirtualMember(r.Context(), familyID, body.Name, callerID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(m)
}

func (h *HouseholdHandler) deleteVirtual(w http.ResponseWriter, r *http.Request) {
	familyID := chi.URLParam(r, "familyID")
	memberID := chi.URLParam(r, "memberID")
	callerID, ok := r.Context().Value(ContextKeyUserID).(string)
	if !ok || callerID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if err := h.families.DeleteVirtualMember(r.Context(), memberID, familyID, callerID); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *HouseholdHandler) updateRole(w http.ResponseWriter, r *http.Request) {
	familyID := chi.URLParam(r, "familyID")
	memberID := chi.URLParam(r, "memberID")
	callerID, ok := r.Context().Value(ContextKeyUserID).(string)
	if !ok || callerID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	var body struct {
		Role string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	if err := h.families.UpdateMemberRole(r.Context(), memberID, familyID, body.Role, callerID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *HouseholdHandler) unlinkedVirtual(w http.ResponseWriter, r *http.Request) {
	familyID := chi.URLParam(r, "familyID")
	if err := requireMember(r, familyID, h.families); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	members, err := h.families.GetUnlinkedVirtualMembers(r.Context(), familyID)
	if err != nil {
		http.Error(w, "failed to get virtual members", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(members)
}

func (h *HouseholdHandler) removeMember(w http.ResponseWriter, r *http.Request) {
	familyID := chi.URLParam(r, "familyID")
	memberID := chi.URLParam(r, "memberID")
	callerID, ok := r.Context().Value(ContextKeyUserID).(string)
	if !ok || callerID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if err := h.families.RemoveMember(r.Context(), memberID, familyID, callerID); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *HouseholdHandler) updateName(w http.ResponseWriter, r *http.Request) {
	familyID := chi.URLParam(r, "familyID")
	callerID, ok := r.Context().Value(ContextKeyUserID).(string)
	if !ok || callerID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	var body struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name == "" {
		http.Error(w, "name required", http.StatusBadRequest)
		return
	}
	if err := h.families.UpdateName(r.Context(), familyID, body.Name, callerID); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *HouseholdHandler) linkVirtual(w http.ResponseWriter, r *http.Request) {
	familyID := chi.URLParam(r, "familyID")
	memberID := chi.URLParam(r, "memberID")
	userID, ok := r.Context().Value(ContextKeyUserID).(string)
	if !ok || userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if err := h.families.LinkVirtualMember(r.Context(), memberID, familyID, userID); err != nil {
		http.Error(w, "failed to link virtual member", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
