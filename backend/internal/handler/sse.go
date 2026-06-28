package handler

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/pidanou/homeboard/internal/service"
)

type SSEHandler struct {
	hub       *Hub
	jwtSecret string
	families  *service.HouseholdService
}

func NewSSEHandler(hub *Hub, jwtSecret string, families *service.HouseholdService) *SSEHandler {
	return &SSEHandler{hub: hub, jwtSecret: jwtSecret, families: families}
}

func (h *SSEHandler) Routes() http.Handler {
	r := chi.NewRouter()
	r.Get("/", h.stream)
	return r
}

func (h *SSEHandler) stream(w http.ResponseWriter, r *http.Request) {
	familyID := chi.URLParam(r, "familyID")

	// EventSource can't set Authorization header; accept token as query param too
	tokenStr := r.URL.Query().Get("token")
	if tokenStr == "" {
		tokenStr = strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	}

	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(h.jwtSecret), nil
	})
	if err != nil || !token.Valid {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	userID, _ := claims["sub"].(string)
	if userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	ctx := context.WithValue(r.Context(), ContextKeyUserID, userID)
	if _, err := h.families.GetMemberRole(ctx, userID, familyID); err != nil {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	ch := h.hub.Subscribe(familyID)
	defer h.hub.Unsubscribe(familyID, ch)

	fmt.Fprintf(w, "data: connected\n\n")
	flusher.Flush()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-ch:
			fmt.Fprintf(w, "data: refresh\n\n")
			flusher.Flush()
		case <-ticker.C:
			fmt.Fprintf(w, ": keepalive\n\n")
			flusher.Flush()
		}
	}
}
