package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/pidanou/family-board/internal/service"
)

type LabelHandler struct {
	labels *service.LabelService
	hub    *Hub
}

func NewLabelHandler(labels *service.LabelService, hub *Hub) *LabelHandler {
	return &LabelHandler{labels: labels, hub: hub}
}

func (h *LabelHandler) Routes() http.Handler {
	r := chi.NewRouter()
	r.Get("/", h.list)
	r.Post("/", h.create)
	r.Delete("/{labelID}", h.delete)
	return r
}

func (h *LabelHandler) list(w http.ResponseWriter, r *http.Request) {
	familyID := chi.URLParam(r, "familyID")
	labels, err := h.labels.ListForFamily(r.Context(), familyID)
	if err != nil {
		http.Error(w, "failed to list labels", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(labels)
}

func (h *LabelHandler) create(w http.ResponseWriter, r *http.Request) {
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
	label, err := h.labels.Create(r.Context(), familyID, body.Name, body.Color)
	if err != nil {
		http.Error(w, "failed to create label", http.StatusInternalServerError)
		return
	}
	h.hub.Broadcast(familyID)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(label)
}

func (h *LabelHandler) delete(w http.ResponseWriter, r *http.Request) {
	familyID := chi.URLParam(r, "familyID")
	labelID := chi.URLParam(r, "labelID")
	if err := h.labels.Delete(r.Context(), labelID, familyID); err != nil {
		http.Error(w, "failed to delete label", http.StatusInternalServerError)
		return
	}
	h.hub.Broadcast(familyID)
	w.WriteHeader(http.StatusNoContent)
}
