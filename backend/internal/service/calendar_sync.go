package service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	ics "github.com/arran4/golang-ical"
	"github.com/google/uuid"
	"github.com/pidanou/homeboard/internal/model"
	"github.com/pidanou/homeboard/internal/repository"
)

const maxFeedBytes = 10 << 20 // 10MB

type CalendarSyncService struct {
	subs   repository.CalendarSubscriptionRepository
	events repository.EventRepository
	client *http.Client
}

func NewCalendarSyncService(subs repository.CalendarSubscriptionRepository, events repository.EventRepository) *CalendarSyncService {
	return &CalendarSyncService{subs: subs, events: events, client: &http.Client{Timeout: 20 * time.Second}}
}

func (s *CalendarSyncService) CreateSubscription(ctx context.Context, familyID, userID, name, url string) (*model.CalendarSubscription, error) {
	sub := &model.CalendarSubscription{
		ID:        uuid.NewString(),
		FamilyID:  familyID,
		Name:      name,
		URL:       url,
		CreatedBy: userID,
		CreatedAt: time.Now().UTC(),
	}
	if err := s.subs.Create(ctx, sub); err != nil {
		return nil, err
	}
	return sub, nil
}

func (s *CalendarSyncService) ListForFamily(ctx context.Context, familyID string) ([]*model.CalendarSubscription, error) {
	return s.subs.ListByFamilyID(ctx, familyID)
}

func (s *CalendarSyncService) Get(ctx context.Context, id string) (*model.CalendarSubscription, error) {
	return s.subs.Get(ctx, id)
}

func (s *CalendarSyncService) Delete(ctx context.Context, id, familyID string) error {
	return s.subs.Delete(ctx, id, familyID)
}

// SyncOne fetches and reconciles a single subscription's feed, recording the outcome.
func (s *CalendarSyncService) SyncOne(ctx context.Context, subscriptionID string) error {
	sub, err := s.subs.Get(ctx, subscriptionID)
	if err != nil {
		return fmt.Errorf("get subscription: %w", err)
	}

	syncErr := s.fetchAndReconcile(ctx, sub)
	var errMsg *string
	if syncErr != nil {
		msg := syncErr.Error()
		errMsg = &msg
	}
	if err := s.subs.UpdateSyncResult(ctx, sub.ID, errMsg); err != nil {
		return fmt.Errorf("record sync result: %w", err)
	}
	return syncErr
}

// SyncAll sweeps every subscription, logging and continuing past per-subscription errors.
// ponytail: single sequential sweep, no concurrency/locking — fine at family-app scale;
// upgrade to per-subscription concurrency only if a sweep is observed exceeding the tick interval.
func (s *CalendarSyncService) SyncAll(ctx context.Context) {
	subs, err := s.subs.ListAll(ctx)
	if err != nil {
		log.Printf("calendar sync: list subscriptions: %v", err)
		return
	}
	for _, sub := range subs {
		if err := s.SyncOne(ctx, sub.ID); err != nil {
			log.Printf("calendar sync: subscription %s (%s): %v", sub.ID, sub.Name, err)
		}
	}
}

func (s *CalendarSyncService) fetchAndReconcile(ctx context.Context, sub *model.CalendarSubscription) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, sub.URL, nil)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("fetch feed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("fetch feed: status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxFeedBytes))
	if err != nil {
		return fmt.Errorf("read feed: %w", err)
	}

	cal, err := ics.ParseCalendar(bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("parse feed: %w", err)
	}

	now := time.Now().UTC()
	keepUIDs := []string{}
	for _, vevent := range cal.Events() {
		uid := vevent.Id()
		title := propValue(vevent, ics.ComponentPropertySummary)
		allDay := isAllDay(vevent)
		start, end, ok := eventTimes(vevent, allDay)
		if uid == "" || title == "" || !ok {
			continue
		}
		keepUIDs = append(keepUIDs, uid)

		var rrule *string
		if r := propValue(vevent, ics.ComponentPropertyRrule); r != "" {
			rrule = &r
		}

		subID := sub.ID
		extUID := uid
		event := &model.Event{
			ID:             uuid.NewString(),
			FamilyID:       sub.FamilyID,
			Title:          title,
			Description:    propValue(vevent, ics.ComponentPropertyDescription),
			Location:       propValue(vevent, ics.ComponentPropertyLocation),
			StartAt:        start,
			EndAt:          end,
			AllDay:         allDay,
			RecurrenceRule: rrule,
			Type:           "default",
			CreatedBy:      sub.CreatedBy,
			CreatedAt:      now,
			UpdatedAt:      now,
			SubscriptionID: &subID,
			ExternalUID:    &extUID,
		}

		id, err := s.events.UpsertExternal(ctx, event)
		if err != nil {
			return fmt.Errorf("upsert event %s: %w", uid, err)
		}

		// Only EXDATE (occurrence cancellation) is honored from the source feed;
		// per-occurrence RECURRENCE-ID overrides are out of scope for v1.
		exdates, _ := vevent.GetExDates()
		for _, occ := range exdates {
			if err := s.events.CancelOccurrence(ctx, id, sub.FamilyID, occ); err != nil {
				return fmt.Errorf("cancel occurrence: %w", err)
			}
		}
	}

	if err := s.events.DeleteStaleExternal(ctx, sub.ID, keepUIDs); err != nil {
		return fmt.Errorf("delete stale events: %w", err)
	}
	return nil
}
