package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/pidanou/family-board/internal/service"
)

type FamilyHandler struct {
	families *service.FamilyService
}

func NewFamilyHandler(families *service.FamilyService) *FamilyHandler {
	return &FamilyHandler{families: families}
}

func (h *FamilyHandler) Routes() http.Handler {
	r := chi.NewRouter()
	r.Get("/", h.list)
	r.Post("/", h.create)
	r.Get("/{familyID}", h.get)
	r.Get("/{familyID}/members", h.members)
	r.Post("/{familyID}/members/virtual", h.createVirtual)
	r.Delete("/{familyID}/members/virtual/{memberID}", h.deleteVirtual)
	r.Post("/{familyID}/members/virtual/{memberID}/link", h.linkVirtual)
	r.Get("/{familyID}/members/virtual/unlinked", h.unlinkedVirtual)
	return r
}

func (h *FamilyHandler) list(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(ContextKeyUserID).(string)

	families, err := h.families.ListForUser(r.Context(), userID)
	if err != nil {
		http.Error(w, "failed to list families", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(families)
}

func (h *FamilyHandler) create(w http.ResponseWriter, r *http.Request) {
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

func (h *FamilyHandler) get(w http.ResponseWriter, r *http.Request) {
	familyID := chi.URLParam(r, "familyID")

	family, err := h.families.GetByID(r.Context(), familyID)
	if err != nil {
		http.Error(w, "family not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(family)
}

func (h *FamilyHandler) members(w http.ResponseWriter, r *http.Request) {
	familyID := chi.URLParam(r, "familyID")

	members, err := h.families.GetMembers(r.Context(), familyID)
	if err != nil {
		http.Error(w, "failed to get members", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(members)
}

func (h *FamilyHandler) createVirtual(w http.ResponseWriter, r *http.Request) {
	familyID := chi.URLParam(r, "familyID")
	var body struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name == "" {
		http.Error(w, "name required", http.StatusBadRequest)
		return
	}
	m, err := h.families.CreateVirtualMember(r.Context(), familyID, body.Name)
	if err != nil {
		http.Error(w, "failed to create virtual member", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(m)
}

func (h *FamilyHandler) deleteVirtual(w http.ResponseWriter, r *http.Request) {
	familyID := chi.URLParam(r, "familyID")
	memberID := chi.URLParam(r, "memberID")
	if err := h.families.DeleteVirtualMember(r.Context(), memberID, familyID); err != nil {
		http.Error(w, "failed to delete virtual member", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *FamilyHandler) unlinkedVirtual(w http.ResponseWriter, r *http.Request) {
	familyID := chi.URLParam(r, "familyID")
	members, err := h.families.GetUnlinkedVirtualMembers(r.Context(), familyID)
	if err != nil {
		http.Error(w, "failed to get virtual members", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(members)
}

func (h *FamilyHandler) linkVirtual(w http.ResponseWriter, r *http.Request) {
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
