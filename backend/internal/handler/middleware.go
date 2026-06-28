package handler

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/pidanou/homeboard/internal/service"
)

type contextKey string

const ContextKeyUserID contextKey = "userID"

func requireAdmin(r *http.Request, familyID string, families *service.HouseholdService) error {
	callerID, ok := r.Context().Value(ContextKeyUserID).(string)
	if !ok || callerID == "" {
		return fmt.Errorf("unauthorized")
	}
	role, err := families.GetMemberRole(r.Context(), callerID, familyID)
	if err != nil || role != "admin" {
		return fmt.Errorf("admin required")
	}
	return nil
}

func requireMember(r *http.Request, familyID string, families *service.HouseholdService) error {
	callerID, ok := r.Context().Value(ContextKeyUserID).(string)
	if !ok || callerID == "" {
		return fmt.Errorf("unauthorized")
	}
	if _, err := families.GetMemberRole(r.Context(), callerID, familyID); err != nil {
		return fmt.Errorf("forbidden")
	}
	return nil
}

// RequireFamilyMember middleware verifies the authenticated user belongs to the family in the URL.
func RequireFamilyMember(families *service.HouseholdService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if err := requireMember(r, chi.URLParam(r, "familyID"), families); err != nil {
				http.Error(w, err.Error(), http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		next.ServeHTTP(w, r)
	})
}

func AuthMiddleware(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if !strings.HasPrefix(authHeader, "Bearer ") {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
			token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(jwtSecret), nil
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

			userID, ok := claims["sub"].(string)
			if !ok {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), ContextKeyUserID, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
