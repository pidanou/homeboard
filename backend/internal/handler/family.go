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
