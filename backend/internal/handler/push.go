package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/pidanou/homeboard/internal/model"
	"github.com/pidanou/homeboard/internal/service"
)

type PushHandler struct {
	push          *service.PushService
	families      *service.HouseholdService
	vapidPublicKey string
}

func NewPushHandler(push *service.PushService, families *service.HouseholdService, vapidPublicKey string) *PushHandler {
	return &PushHandler{push: push, families: families, vapidPublicKey: vapidPublicKey}
}

// PublicRoutes returns routes that don't require auth (VAPID public key).
func (h *PushHandler) PublicRoutes() http.Handler {
	r := chi.NewRouter()
	r.Get("/vapid-public-key", h.vapidKey)
	return r
}

// Routes returns family-scoped push routes (require auth + membership via middleware).
func (h *PushHandler) Routes() http.Handler {
	r := chi.NewRouter()
	r.Post("/subscribe", h.subscribe)
	r.Delete("/subscribe", h.unsubscribe)
	return r
}

func (h *PushHandler) vapidKey(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"public_key": h.vapidPublicKey})
}

func (h *PushHandler) subscribe(w http.ResponseWriter, r *http.Request) {
	familyID := chi.URLParam(r, "familyID")
	userID := r.Context().Value(ContextKeyUserID).(string)

	var body struct {
		Endpoint string `json:"endpoint"`
		Auth     string `json:"auth"`
		P256DH   string `json:"p256dh"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Endpoint == "" {
		http.Error(w, "invalid subscription", http.StatusBadRequest)
		return
	}

	if err := h.push.Subscribe(r.Context(), &model.PushSubscription{
		UserID:   userID,
		FamilyID: familyID,
		Endpoint: body.Endpoint,
		Auth:     body.Auth,
		P256DH:   body.P256DH,
	}); err != nil {
		http.Error(w, "failed to save subscription", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *PushHandler) unsubscribe(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(ContextKeyUserID).(string)

	var body struct {
		Endpoint string `json:"endpoint"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Endpoint == "" {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if err := h.push.Unsubscribe(r.Context(), userID, body.Endpoint); err != nil {
		http.Error(w, "failed to remove subscription", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
