package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/pidanou/homeboard/internal/model"
	"github.com/pidanou/homeboard/internal/repository"
)

type EventService struct {
	events repository.EventRepository
}

func NewEventService(events repository.EventRepository) *EventService {
	return &EventService{events: events}
}

func (s *EventService) Create(ctx context.Context, familyID, userID, title, description, location string, startAt, endAt time.Time, allDay bool, attendeeIDs []string, categoryID *string, recurrenceRule *string, eventType string, icon *string, birthdayOf *string, important bool) (*model.Event, error) {
	if eventType == "" {
		eventType = "default"
	}
	now := time.Now().UTC()
	event := &model.Event{
		ID:             uuid.NewString(),
		FamilyID:       familyID,
		Title:          title,
		Description:    description,
		Location:       location,
		StartAt:        startAt.UTC(),
		EndAt:          endAt.UTC(),
		AllDay:         allDay,
		AttendeeIDs:    attendeeIDs,
		CategoryID:     categoryID,
		RecurrenceRule: recurrenceRule,
		Type:           eventType,
		Icon:           icon,
		BirthdayOf:     birthdayOf,
		Important:      important,
		CreatedBy:      userID,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if err := s.events.Create(ctx, event); err != nil {
		return nil, err
	}
	return event, nil
}

func (s *EventService) ListForRange(ctx context.Context, familyID string, from, to time.Time) ([]*model.Event, error) {
	return s.events.ListByFamilyAndRange(ctx, familyID, from, to)
}

// Update updates the parent event (all occurrences).
func (s *EventService) Update(ctx context.Context, event *model.Event) error {
	event.UpdatedAt = time.Now().UTC()
	return s.events.Update(ctx, event)
}

// UpdateOccurrence creates an exception row for a single occurrence.
func (s *EventService) UpdateOccurrence(ctx context.Context, parentID, familyID, userID string, occDate time.Time, event *model.Event) error {
	now := time.Now().UTC()
	event.ID = fmt.Sprintf("%s::%s", parentID, occDate.Format("20060102"))
	event.FamilyID = familyID
	event.RecurrenceParentID = &parentID
	event.RecurrenceDate = &occDate
	event.CreatedBy = userID
	event.CreatedAt = now
	event.UpdatedAt = now
	return s.events.CreateException(ctx, event)
}

func (s *EventService) Delete(ctx context.Context, eventID, familyID string) error {
	return s.events.Delete(ctx, eventID, familyID)
}

func (s *EventService) CancelOccurrence(ctx context.Context, parentID, familyID string, date time.Time) error {
	return s.events.CancelOccurrence(ctx, parentID, familyID, date)
}
