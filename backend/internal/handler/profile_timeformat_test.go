package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/pidanou/homeboard/internal/model"
)

func TestUpdateTimeFormat(t *testing.T) {
	e, userID := newProfileTestEnv(t)

	t.Run("rejects unsupported time format", func(t *testing.T) {
		resp := e.do("PATCH", "/profile/time-format", e.token(userID), map[string]string{"time_format": "30"})
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("want 400, got %d", resp.StatusCode)
		}
	})

	t.Run("persists supported time format", func(t *testing.T) {
		resp := e.do("PATCH", "/profile/time-format", e.token(userID), map[string]string{"time_format": "24"})
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("want 200, got %d", resp.StatusCode)
		}
		var user model.User
		json.NewDecoder(resp.Body).Decode(&user)
		if user.TimeFormat != "24" {
			t.Errorf("want time format 24, got %q", user.TimeFormat)
		}

		var stored string
		e.pool.QueryRow(context.Background(), `SELECT time_format FROM users WHERE id = $1`, userID).Scan(&stored)
		if stored != "24" {
			t.Errorf("db want 24, got %q", stored)
		}
	})
}

func TestNewUserDefaultsToAutoTimeFormat(t *testing.T) {
	e, userID := newProfileTestEnv(t)

	resp := e.do("GET", "/profile", e.token(userID), nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("want 200, got %d", resp.StatusCode)
	}
	var user model.User
	json.NewDecoder(resp.Body).Decode(&user)
	if user.TimeFormat != "auto" {
		t.Errorf("want default time format auto, got %q", user.TimeFormat)
	}
}
