package service

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pidanou/homeboard/internal/repository/postgres"
)

const feedTwoEvents = `BEGIN:VCALENDAR
VERSION:2.0
PRODID:-//test//test//EN
BEGIN:VEVENT
UID:event-1@example.com
DTSTART:20260801T100000Z
DTEND:20260801T110000Z
SUMMARY:Feed Event One
END:VEVENT
BEGIN:VEVENT
UID:event-2@example.com
DTSTART:20260802T100000Z
DTEND:20260802T110000Z
SUMMARY:Feed Event Two
END:VEVENT
END:VCALENDAR
`

const feedOneEvent = `BEGIN:VCALENDAR
VERSION:2.0
PRODID:-//test//test//EN
BEGIN:VEVENT
UID:event-1@example.com
DTSTART:20260801T100000Z
DTEND:20260801T110000Z
SUMMARY:Feed Event One Updated
END:VEVENT
END:VCALENDAR
`

func newSyncTestPool(t *testing.T) *pgxpool.Pool {
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
	t.Cleanup(pool.Close)
	return pool
}

func seedSyncFamily(t *testing.T, pool *pgxpool.Pool) (familyID, userID string) {
	t.Helper()
	ctx := context.Background()
	familyID = uuid.NewString()
	userID = uuid.NewString()

	if _, err := pool.Exec(ctx,
		`INSERT INTO users (id, email, name, password_hash, created_at) VALUES ($1, $2, $3, 'x', $4)`,
		userID, fmt.Sprintf("sync-%s@test.com", userID[:8]), "Sync Test User", time.Now().UTC(),
	); err != nil {
		t.Fatalf("insert user: %v", err)
	}
	if _, err := pool.Exec(ctx,
		`INSERT INTO households (id, name, created_at) VALUES ($1, $2, $3)`,
		familyID, "Sync Test Family", time.Now().UTC(),
	); err != nil {
		t.Fatalf("insert household: %v", err)
	}

	t.Cleanup(func() {
		pool.Exec(ctx, `DELETE FROM households WHERE id = $1`, familyID)
		pool.Exec(ctx, `DELETE FROM users WHERE id = $1`, userID)
	})

	return familyID, userID
}

// TestCalendarSyncReconciliation exercises the actual reconciliation risk in subscription
// sync: a second sync must update existing rows and delete events whose UID disappeared
// from the feed, not just append duplicates.
func TestCalendarSyncReconciliation(t *testing.T) {
	pool := newSyncTestPool(t)
	familyID, userID := seedSyncFamily(t, pool)

	eventRepo := postgres.NewEventRepository(pool)
	subRepo := postgres.NewCalendarSubscriptionRepository(pool)
	syncSvc := NewCalendarSyncService(subRepo, eventRepo)

	body := feedTwoEvents
	feed := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/calendar")
		w.Write([]byte(body))
	}))
	defer feed.Close()

	sub, err := syncSvc.CreateSubscription(context.Background(), familyID, userID, "Test Feed", feed.URL)
	if err != nil {
		t.Fatalf("create subscription: %v", err)
	}
	t.Cleanup(func() {
		pool.Exec(context.Background(), `DELETE FROM calendar_subscriptions WHERE id = $1`, sub.ID)
	})

	if err := syncSvc.SyncOne(context.Background(), sub.ID); err != nil {
		t.Fatalf("first sync: %v", err)
	}

	events, err := eventRepo.ListAllForExport(context.Background(), familyID)
	if err != nil {
		t.Fatalf("list events: %v", err)
	}
	// ListAllForExport excludes synced events by design, so check directly.
	var count int
	if err := pool.QueryRow(context.Background(),
		`SELECT count(*) FROM events WHERE subscription_id = $1`, sub.ID).Scan(&count); err != nil {
		t.Fatalf("count events: %v", err)
	}
	if count != 2 {
		t.Fatalf("want 2 synced events after first sync, got %d", count)
	}
	_ = events

	// Second sync: feed now has one event (updated) and is missing the other — the
	// stale one must be deleted, the remaining one must be updated in place.
	body = feedOneEvent
	if err := syncSvc.SyncOne(context.Background(), sub.ID); err != nil {
		t.Fatalf("second sync: %v", err)
	}

	var titles []string
	rows, err := pool.Query(context.Background(),
		`SELECT title FROM events WHERE subscription_id = $1`, sub.ID)
	if err != nil {
		t.Fatalf("query events: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var title string
		if err := rows.Scan(&title); err != nil {
			t.Fatalf("scan: %v", err)
		}
		titles = append(titles, title)
	}
	if len(titles) != 1 {
		t.Fatalf("want 1 event after second sync (stale deleted), got %d: %v", len(titles), titles)
	}
	if titles[0] != "Feed Event One Updated" {
		t.Errorf("want updated title, got %q", titles[0])
	}
}
