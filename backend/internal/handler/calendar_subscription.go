package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/pidanou/homeboard/internal/service"
)

type CalendarSubscriptionHandler struct {
	sync     *service.CalendarSyncService
	families *service.HouseholdService
}

func NewCalendarSubscriptionHandler(sync *service.CalendarSyncService, families *service.HouseholdService) *CalendarSubscriptionHandler {
	return &CalendarSubscriptionHandler{sync: sync, families: families}
}

// Routes is mounted at /households/{familyID}/calendar/subscriptions, authenticated + admin-gated.
func (h *CalendarSubscriptionHandler) Routes() http.Handler {
	r := chi.NewRouter()
	r.Get("/", h.list)
	r.Post("/", h.create)
	r.Delete("/{id}", h.delete)
	r.Post("/{id}/sync", h.syncNow)
	return r
}

func (h *CalendarSubscriptionHandler) list(w http.ResponseWriter, r *http.Request) {
	familyID := chi.URLParam(r, "familyID")
	if err := requireAdmin(r, familyID, h.families); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	subs, err := h.sync.ListForFamily(r.Context(), familyID)
	if err != nil {
		http.Error(w, "failed to list subscriptions", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(subs)
}

func (h *CalendarSubscriptionHandler) create(w http.ResponseWriter, r *http.Request) {
	familyID := chi.URLParam(r, "familyID")
	if err := requireAdmin(r, familyID, h.families); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	userID := r.Context().Value(ContextKeyUserID).(string)

	var body struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name == "" || body.URL == "" {
		http.Error(w, "name and url are required", http.StatusBadRequest)
		return
	}

	sub, err := h.sync.CreateSubscription(r.Context(), familyID, userID, body.Name, body.URL)
	if err != nil {
		http.Error(w, "failed to create subscription", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(sub)
}

func (h *CalendarSubscriptionHandler) delete(w http.ResponseWriter, r *http.Request) {
	familyID := chi.URLParam(r, "familyID")
	if err := requireAdmin(r, familyID, h.families); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	id := chi.URLParam(r, "id")
	if err := h.sync.Delete(r.Context(), id, familyID); err != nil {
		http.Error(w, "failed to delete subscription", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *CalendarSubscriptionHandler) syncNow(w http.ResponseWriter, r *http.Request) {
	familyID := chi.URLParam(r, "familyID")
	if err := requireAdmin(r, familyID, h.families); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	id := chi.URLParam(r, "id")

	sub, err := h.sync.Get(r.Context(), id)
	if err != nil || sub.FamilyID != familyID {
		http.Error(w, "subscription not found", http.StatusNotFound)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 25*time.Second)
	defer cancel()
	if err := h.sync.SyncOne(ctx, id); err != nil {
		http.Error(w, "sync failed: "+err.Error(), http.StatusBadGateway)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
