package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/pidanou/family-board/internal/model"
	"github.com/pidanou/family-board/internal/service"
)

type EventHandler struct {
	events *service.EventService
	hub    *Hub
}

func NewEventHandler(events *service.EventService, hub *Hub) *EventHandler {
	return &EventHandler{events: events, hub: hub}
}

func (h *EventHandler) Routes() http.Handler {
	r := chi.NewRouter()
	r.Get("/", h.list)
	r.Post("/", h.create)
	r.Patch("/{eventID}", h.update)
	r.Delete("/{eventID}", h.delete)
	return r
}

func (h *EventHandler) list(w http.ResponseWriter, r *http.Request) {
	familyID := chi.URLParam(r, "familyID")

	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")

	from, err := time.Parse(time.RFC3339, fromStr)
	if err != nil {
		http.Error(w, "invalid from date", http.StatusBadRequest)
		return
	}
	to, err := time.Parse(time.RFC3339, toStr)
	if err != nil {
		http.Error(w, "invalid to date", http.StatusBadRequest)
		return
	}

	events, err := h.events.ListForRange(r.Context(), familyID, from, to)
	if err != nil {
		http.Error(w, "failed to list events", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(events)
}

func (h *EventHandler) create(w http.ResponseWriter, r *http.Request) {
	familyID := chi.URLParam(r, "familyID")
	userID := r.Context().Value(ContextKeyUserID).(string)

	var body struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		StartAt     string `json:"start_at"`
		EndAt       string `json:"end_at"`
		AllDay      bool   `json:"all_day"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Title == "" {
		http.Error(w, "title is required", http.StatusBadRequest)
		return
	}

	startAt, err := time.Parse(time.RFC3339, body.StartAt)
	if err != nil {
		http.Error(w, "invalid start_at", http.StatusBadRequest)
		return
	}
	endAt, err := time.Parse(time.RFC3339, body.EndAt)
	if err != nil {
		http.Error(w, "invalid end_at", http.StatusBadRequest)
		return
	}

	event, err := h.events.Create(r.Context(), familyID, userID, body.Title, body.Description, startAt, endAt, body.AllDay)
	if err != nil {
		http.Error(w, "failed to create event", http.StatusInternalServerError)
		return
	}

	h.hub.Broadcast(familyID)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(event)
}

func (h *EventHandler) update(w http.ResponseWriter, r *http.Request) {
	eventID := chi.URLParam(r, "eventID")

	var body struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		StartAt     string `json:"start_at"`
		EndAt       string `json:"end_at"`
		AllDay      bool   `json:"all_day"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	startAt, err := time.Parse(time.RFC3339, body.StartAt)
	if err != nil {
		http.Error(w, "invalid start_at", http.StatusBadRequest)
		return
	}
	endAt, err := time.Parse(time.RFC3339, body.EndAt)
	if err != nil {
		http.Error(w, "invalid end_at", http.StatusBadRequest)
		return
	}

	familyID := chi.URLParam(r, "familyID")
	event := &model.Event{
		ID:          eventID,
		FamilyID:    familyID,
		Title:       body.Title,
		Description: body.Description,
		StartAt:     startAt,
		EndAt:       endAt,
		AllDay:      body.AllDay,
	}

	if err := h.events.Update(r.Context(), event); err != nil {
		http.Error(w, "failed to update event", http.StatusInternalServerError)
		return
	}

	h.hub.Broadcast(familyID)
	w.WriteHeader(http.StatusNoContent)
}

func (h *EventHandler) delete(w http.ResponseWriter, r *http.Request) {
	familyID := chi.URLParam(r, "familyID")
	eventID := chi.URLParam(r, "eventID")

	if err := h.events.Delete(r.Context(), eventID, familyID); err != nil {
		http.Error(w, "failed to delete event", http.StatusInternalServerError)
		return
	}

	h.hub.Broadcast(familyID)
	w.WriteHeader(http.StatusNoContent)
}
