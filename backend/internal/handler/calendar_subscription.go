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
		writeError(w, http.StatusForbidden, codeFor(err, "forbidden"))
		return
	}
	subs, err := h.sync.ListForFamily(r.Context(), familyID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "list_subscriptions_failed")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(subs)
}

func (h *CalendarSubscriptionHandler) create(w http.ResponseWriter, r *http.Request) {
	familyID := chi.URLParam(r, "familyID")
	if err := requireAdmin(r, familyID, h.families); err != nil {
		writeError(w, http.StatusForbidden, codeFor(err, "forbidden"))
		return
	}
	userID := r.Context().Value(ContextKeyUserID).(string)

	var body struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name == "" || body.URL == "" {
		writeError(w, http.StatusBadRequest, "name_and_url_required")
		return
	}

	sub, err := h.sync.CreateSubscription(r.Context(), familyID, userID, body.Name, body.URL)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "create_subscription_failed")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(sub)
}

func (h *CalendarSubscriptionHandler) delete(w http.ResponseWriter, r *http.Request) {
	familyID := chi.URLParam(r, "familyID")
	if err := requireAdmin(r, familyID, h.families); err != nil {
		writeError(w, http.StatusForbidden, codeFor(err, "forbidden"))
		return
	}
	id := chi.URLParam(r, "id")
	if err := h.sync.Delete(r.Context(), id, familyID); err != nil {
		writeError(w, http.StatusInternalServerError, "delete_subscription_failed")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *CalendarSubscriptionHandler) syncNow(w http.ResponseWriter, r *http.Request) {
	familyID := chi.URLParam(r, "familyID")
	if err := requireAdmin(r, familyID, h.families); err != nil {
		writeError(w, http.StatusForbidden, codeFor(err, "forbidden"))
		return
	}
	id := chi.URLParam(r, "id")

	sub, err := h.sync.Get(r.Context(), id)
	if err != nil || sub.FamilyID != familyID {
		writeError(w, http.StatusNotFound, "subscription_not_found")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 25*time.Second)
	defer cancel()
	if err := h.sync.SyncOne(ctx, id); err != nil {
		writeError(w, http.StatusBadGateway, "sync_failed")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
