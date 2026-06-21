package handler

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/pidanou/homeboard/internal/model"
	"github.com/pidanou/homeboard/internal/service"
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

// parseVirtualID splits "parentID::YYYYMMDD" into (parentID, &date) or (id, nil).
func parseVirtualID(raw string) (string, *time.Time) {
	parts := strings.SplitN(raw, "::", 2)
	if len(parts) == 2 {
		t, err := time.Parse("20060102", parts[1])
		if err == nil {
			return parts[0], &t
		}
	}
	return raw, nil
}

func (h *EventHandler) list(w http.ResponseWriter, r *http.Request) {
	familyID := chi.URLParam(r, "familyID")
	from, err := time.Parse(time.RFC3339, r.URL.Query().Get("from"))
	if err != nil {
		http.Error(w, "invalid from date", http.StatusBadRequest)
		return
	}
	to, err := time.Parse(time.RFC3339, r.URL.Query().Get("to"))
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
		Title          string   `json:"title"`
		Description    string   `json:"description"`
		Location       string   `json:"location"`
		StartAt        string   `json:"start_at"`
		EndAt          string   `json:"end_at"`
		AllDay         bool     `json:"all_day"`
		AttendeeIDs    []string `json:"attendee_ids"`
		CategoryID     *string  `json:"category_id"`
		RecurrenceRule *string  `json:"recurrence_rule"`
		Type           string   `json:"type"`
		Icon           *string  `json:"icon"`
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

	event, err := h.events.Create(r.Context(), familyID, userID, body.Title, body.Description, body.Location, startAt, endAt, body.AllDay, body.AttendeeIDs, body.CategoryID, body.RecurrenceRule, body.Type, body.Icon)
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
	rawID := chi.URLParam(r, "eventID")
	familyID := chi.URLParam(r, "familyID")
	userID := r.Context().Value(ContextKeyUserID).(string)

	var body struct {
		Title          string   `json:"title"`
		Description    string   `json:"description"`
		Location       string   `json:"location"`
		StartAt        string   `json:"start_at"`
		EndAt          string   `json:"end_at"`
		AllDay         bool     `json:"all_day"`
		AttendeeIDs    []string `json:"attendee_ids"`
		CategoryID     *string  `json:"category_id"`
		RecurrenceRule *string  `json:"recurrence_rule"`
		Icon           *string  `json:"icon"`
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

	parentID, occDate := parseVirtualID(rawID)

	if occDate != nil {
		// Edit this occurrence only → create exception row.
		exc := &model.Event{
			Title: body.Title, Description: body.Description, Location: body.Location,
			StartAt: startAt.UTC(), EndAt: endAt.UTC(), AllDay: body.AllDay,
			AttendeeIDs: body.AttendeeIDs, CategoryID: body.CategoryID,
		}
		if err := h.events.UpdateOccurrence(r.Context(), parentID, familyID, userID, *occDate, exc); err != nil {
			http.Error(w, "failed to update occurrence", http.StatusInternalServerError)
			return
		}
	} else {
		event := &model.Event{
			ID: parentID, FamilyID: familyID,
			Title: body.Title, Description: body.Description, Location: body.Location,
			StartAt: startAt.UTC(), EndAt: endAt.UTC(), AllDay: body.AllDay,
			AttendeeIDs: body.AttendeeIDs, CategoryID: body.CategoryID,
			RecurrenceRule: body.RecurrenceRule, Icon: body.Icon,
		}
		if err := h.events.Update(r.Context(), event); err != nil {
			http.Error(w, "failed to update event", http.StatusInternalServerError)
			return
		}
	}

	h.hub.Broadcast(familyID)
	w.WriteHeader(http.StatusNoContent)
}

func (h *EventHandler) delete(w http.ResponseWriter, r *http.Request) {
	familyID := chi.URLParam(r, "familyID")
	rawID := chi.URLParam(r, "eventID")

	parentID, occDate := parseVirtualID(rawID)

	if occDate != nil {
		if err := h.events.CancelOccurrence(r.Context(), parentID, familyID, *occDate); err != nil {
			http.Error(w, "failed to cancel occurrence", http.StatusInternalServerError)
			return
		}
	} else {
		if err := h.events.Delete(r.Context(), parentID, familyID); err != nil {
			http.Error(w, "failed to delete event", http.StatusInternalServerError)
			return
		}
	}

	h.hub.Broadcast(familyID)
	w.WriteHeader(http.StatusNoContent)
}
