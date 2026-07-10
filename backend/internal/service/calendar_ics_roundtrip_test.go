package service

import (
	"context"
	"testing"
	"time"

	"github.com/pidanou/homeboard/internal/repository/postgres"
)

// TestCalendarICSRoundTrip builds an event, exports it to ICS, then re-imports that ICS
// into a second household — this is the actual risk in export/import: RFC5545 formatting
// (line folding, DATE-TIME) must survive both a write and a read.
func TestCalendarICSRoundTrip(t *testing.T) {
	pool := newSyncTestPool(t)
	sourceFamilyID, userID := seedSyncFamily(t, pool)
	destFamilyID, _ := seedSyncFamily(t, pool)

	eventRepo := postgres.NewEventRepository(pool)
	tokenRepo := postgres.NewCalendarExportTokenRepository(pool)
	householdRepo := postgres.NewHouseholdRepository(pool)

	eventSvc := NewEventService(eventRepo)
	exportSvc := NewCalendarExportService(tokenRepo, eventRepo, householdRepo)
	importSvc := NewCalendarImportService(eventSvc)

	start := time.Date(2026, 8, 15, 14, 30, 0, 0, time.UTC)
	end := start.Add(time.Hour)
	created, err := eventSvc.Create(context.Background(), sourceFamilyID, userID,
		"Round Trip Meeting", "desc", "Somewhere", start, end, false, nil, nil, nil, "default", nil, false)
	if err != nil {
		t.Fatalf("create event: %v", err)
	}
	t.Cleanup(func() {
		pool.Exec(context.Background(), `DELETE FROM events WHERE id = $1`, created.ID)
	})

	token, err := exportSvc.Regenerate(context.Background(), sourceFamilyID, userID)
	if err != nil {
		t.Fatalf("regenerate export token: %v", err)
	}

	data, err := exportSvc.BuildICS(context.Background(), token.Token)
	if err != nil {
		t.Fatalf("build ics: %v", err)
	}

	imported, skipped, err := importSvc.ImportFile(context.Background(), destFamilyID, userID, data)
	if err != nil {
		t.Fatalf("import ics: %v", err)
	}
	if imported != 1 || skipped != 0 {
		t.Fatalf("want 1 imported, 0 skipped; got %d imported, %d skipped", imported, skipped)
	}

	events, err := eventRepo.ListAllForExport(context.Background(), destFamilyID)
	if err != nil {
		t.Fatalf("list dest events: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("want 1 event in dest household, got %d", len(events))
	}
	got := events[0]
	t.Cleanup(func() {
		pool.Exec(context.Background(), `DELETE FROM events WHERE id = $1`, got.ID)
	})
	if got.Title != "Round Trip Meeting" {
		t.Errorf("want title survived, got %q", got.Title)
	}
	if !got.StartAt.Equal(start) {
		t.Errorf("want start %v, got %v", start, got.StartAt)
	}
	if !got.EndAt.Equal(end) {
		t.Errorf("want end %v, got %v", end, got.EndAt)
	}
}
