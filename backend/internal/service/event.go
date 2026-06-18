package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/pidanou/family-board/internal/model"
	"github.com/pidanou/family-board/internal/repository"
)

type EventService struct {
	events repository.EventRepository
}

func NewEventService(events repository.EventRepository) *EventService {
	return &EventService{events: events}
}

func (s *EventService) Create(ctx context.Context, familyID, userID, title, description, location string, startAt, endAt time.Time, allDay bool, attendeeIDs []string, categoryID *string) (*model.Event, error) {
	now := time.Now().UTC()
	event := &model.Event{
		ID:          uuid.NewString(),
		FamilyID:    familyID,
		Title:       title,
		Description: description,
		Location:    location,
		StartAt:     startAt.UTC(),
		EndAt:       endAt.UTC(),
		AllDay:      allDay,
		AttendeeIDs: attendeeIDs,
		CategoryID:  categoryID,
		CreatedBy:   userID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.events.Create(ctx, event); err != nil {
		return nil, err
	}
	return event, nil
}

func (s *EventService) ListForRange(ctx context.Context, familyID string, from, to time.Time) ([]*model.Event, error) {
	return s.events.ListByFamilyAndRange(ctx, familyID, from, to)
}

func (s *EventService) Update(ctx context.Context, event *model.Event) error {
	event.UpdatedAt = time.Now().UTC()
	return s.events.Update(ctx, event)
}

func (s *EventService) Delete(ctx context.Context, eventID, familyID string) error {
	return s.events.Delete(ctx, eventID, familyID)
}
