package handler_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/pidanou/homeboard/internal/model"
)

func TestCalendarExportToken(t *testing.T) {
	e := newTestEnv(t)
	familyID, adminID, memberID := e.seedFamily(t)
	url := fmt.Sprintf("/households/%s/calendar/export-token", familyID)

	t.Run("member cannot regenerate", func(t *testing.T) {
		resp := e.do("POST", url, e.token(memberID), nil)
		if resp.StatusCode != http.StatusForbidden {
			t.Errorf("want 403, got %d", resp.StatusCode)
		}
	})

	var tok model.CalendarExportToken
	t.Run("admin can regenerate", func(t *testing.T) {
		resp := e.do("POST", url, e.token(adminID), nil)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("want 200, got %d", resp.StatusCode)
		}
		json.NewDecoder(resp.Body).Decode(&tok)
		if tok.Token == "" {
			t.Fatal("expected non-empty token")
		}
	})

	t.Run("public ics endpoint serves with valid token", func(t *testing.T) {
		resp, err := http.Get(e.server.URL + "/calendar/export/" + tok.Token + ".ics")
		if err != nil {
			t.Fatalf("get: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("want 200, got %d", resp.StatusCode)
		}
		if ct := resp.Header.Get("Content-Type"); ct != "text/calendar; charset=utf-8" {
			t.Errorf("want text/calendar content-type, got %q", ct)
		}
	})

	t.Run("public ics endpoint 404s on invalid token", func(t *testing.T) {
		resp, err := http.Get(e.server.URL + "/calendar/export/nonexistent.ics")
		if err != nil {
			t.Fatalf("get: %v", err)
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("want 404, got %d", resp.StatusCode)
		}
	})

	t.Run("member cannot revoke", func(t *testing.T) {
		resp := e.do("DELETE", url, e.token(memberID), nil)
		if resp.StatusCode != http.StatusForbidden {
			t.Errorf("want 403, got %d", resp.StatusCode)
		}
	})

	t.Run("admin can revoke", func(t *testing.T) {
		resp := e.do("DELETE", url, e.token(adminID), nil)
		if resp.StatusCode != http.StatusNoContent {
			t.Errorf("want 204, got %d", resp.StatusCode)
		}
	})
}
