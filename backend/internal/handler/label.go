package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/pidanou/family-board/internal/service"
)

type CategoryHandler struct {
	categories *service.CategoryService
	hub        *Hub
}

func NewCategoryHandler(categories *service.CategoryService, hub *Hub) *CategoryHandler {
	return &CategoryHandler{categories: categories, hub: hub}
}

func (h *CategoryHandler) Routes() http.Handler {
	r := chi.NewRouter()
	r.Get("/", h.list)
	r.Post("/", h.create)
	r.Delete("/{categoryID}", h.delete)
	return r
}

func (h *CategoryHandler) list(w http.ResponseWriter, r *http.Request) {
	familyID := chi.URLParam(r, "familyID")
	categories, err := h.categories.ListForFamily(r.Context(), familyID)
	if err != nil {
		http.Error(w, "failed to list categories", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(categories)
}

func (h *CategoryHandler) create(w http.ResponseWriter, r *http.Request) {
	familyID := chi.URLParam(r, "familyID")
	var body struct {
		Name  string `json:"name"`
		Color string `json:"color"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}
	if body.Color == "" {
		body.Color = "gray"
	}
	category, err := h.categories.Create(r.Context(), familyID, body.Name, body.Color)
	if err != nil {
		http.Error(w, "failed to create category", http.StatusInternalServerError)
		return
	}
	h.hub.Broadcast(familyID)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(category)
}

func (h *CategoryHandler) delete(w http.ResponseWriter, r *http.Request) {
	familyID := chi.URLParam(r, "familyID")
	categoryID := chi.URLParam(r, "categoryID")
	if err := h.categories.Delete(r.Context(), categoryID, familyID); err != nil {
		http.Error(w, "failed to delete category", http.StatusInternalServerError)
		return
	}
	h.hub.Broadcast(familyID)
	w.WriteHeader(http.StatusNoContent)
}
