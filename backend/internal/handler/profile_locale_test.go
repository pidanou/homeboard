package handler_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pidanou/homeboard/internal/handler"
	"github.com/pidanou/homeboard/internal/model"
	pgRepo "github.com/pidanou/homeboard/internal/repository/postgres"
	"github.com/pidanou/homeboard/internal/service"
)

func newProfileTestEnv(t *testing.T) (*testEnv, string) {
	t.Helper()
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		dbURL = os.Getenv("DATABASE_URL")
	}
	if dbURL == "" {
		t.Skip("DATABASE_URL not set — skipping integration tests")
	}

	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		t.Fatalf("connect db: %v", err)
	}

	userID := uuid.NewString()
	_, err = pool.Exec(context.Background(),
		`INSERT INTO users (id, email, name, password_hash, created_at) VALUES ($1, $2, $3, 'x', $4)`,
		userID, fmt.Sprintf("locale-%s@test.com", userID[:8]), "Locale Test", time.Now().UTC())
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}

	userRepo := pgRepo.NewUserRepository(pool)
	authSvc := service.NewAuthService(userRepo, testJWTSecret, nil)
	profileH := handler.NewProfileHandler(authSvc, t.TempDir(), "http://localhost")

	r := chi.NewRouter()
	r.Group(func(r chi.Router) {
		r.Use(handler.AuthMiddleware(testJWTSecret))
		r.Mount("/profile", profileH.Routes())
	})

	srv := httptest.NewServer(r)
	t.Cleanup(func() {
		srv.Close()
		pool.Exec(context.Background(), `DELETE FROM users WHERE id = $1`, userID)
		pool.Close()
	})
	return &testEnv{server: srv, pool: pool}, userID
}

func TestUpdateLocale(t *testing.T) {
	e, userID := newProfileTestEnv(t)

	t.Run("rejects unsupported locale", func(t *testing.T) {
		resp := e.do("PATCH", "/profile/locale", e.token(userID), map[string]string{"locale": "de"})
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("want 400, got %d", resp.StatusCode)
		}
	})

	t.Run("persists supported locale", func(t *testing.T) {
		resp := e.do("PATCH", "/profile/locale", e.token(userID), map[string]string{"locale": "fr"})
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("want 200, got %d", resp.StatusCode)
		}
		var user model.User
		json.NewDecoder(resp.Body).Decode(&user)
		if user.Locale != "fr" {
			t.Errorf("want locale fr, got %q", user.Locale)
		}

		var stored string
		e.pool.QueryRow(context.Background(), `SELECT locale FROM users WHERE id = $1`, userID).Scan(&stored)
		if stored != "fr" {
			t.Errorf("db want fr, got %q", stored)
		}
	})
}

func TestNewUserDefaultsToEnglishLocale(t *testing.T) {
	e, userID := newProfileTestEnv(t)

	resp := e.do("GET", "/profile", e.token(userID), nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("want 200, got %d", resp.StatusCode)
	}
	var user model.User
	json.NewDecoder(resp.Body).Decode(&user)
	if user.Locale != "en" {
		t.Errorf("want default locale en, got %q", user.Locale)
	}
}
