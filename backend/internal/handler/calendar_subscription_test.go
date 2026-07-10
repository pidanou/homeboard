package handler_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/pidanou/homeboard/internal/model"
)

func TestCalendarSubscriptionMutations(t *testing.T) {
	e := newTestEnv(t)
	familyID, adminID, memberID := e.seedFamily(t)
	url := fmt.Sprintf("/households/%s/calendar/subscriptions", familyID)
	body := map[string]string{"name": "Test Feed", "url": "https://example.com/feed.ics"}

	t.Run("member cannot create subscription", func(t *testing.T) {
		resp := e.do("POST", url, e.token(memberID), body)
		if resp.StatusCode != http.StatusForbidden {
			t.Errorf("want 403, got %d", resp.StatusCode)
		}
	})

	var sub model.CalendarSubscription
	t.Run("admin can create subscription", func(t *testing.T) {
		resp := e.do("POST", url, e.token(adminID), body)
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("want 201, got %d", resp.StatusCode)
		}
		json.NewDecoder(resp.Body).Decode(&sub)
		if sub.ID == "" {
			t.Fatal("expected non-empty subscription id")
		}
	})

	t.Run("admin can list subscriptions", func(t *testing.T) {
		resp := e.do("GET", url, e.token(adminID), nil)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("want 200, got %d", resp.StatusCode)
		}
		var subs []model.CalendarSubscription
		json.NewDecoder(resp.Body).Decode(&subs)
		if len(subs) != 1 {
			t.Errorf("want 1 subscription, got %d", len(subs))
		}
	})

	t.Run("member cannot delete subscription", func(t *testing.T) {
		resp := e.do("DELETE", fmt.Sprintf("%s/%s", url, sub.ID), e.token(memberID), nil)
		if resp.StatusCode != http.StatusForbidden {
			t.Errorf("want 403, got %d", resp.StatusCode)
		}
	})

	t.Run("admin can delete subscription", func(t *testing.T) {
		resp := e.do("DELETE", fmt.Sprintf("%s/%s", url, sub.ID), e.token(adminID), nil)
		if resp.StatusCode != http.StatusNoContent {
			t.Errorf("want 204, got %d", resp.StatusCode)
		}
	})
}
