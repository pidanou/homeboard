package repository

import (
	"context"
	"time"

	"github.com/pidanou/family-board/internal/model"
)

type EventRepository interface {
	Create(ctx context.Context, event *model.Event) error
	ListByFamilyAndRange(ctx context.Context, familyID string, from, to time.Time) ([]*model.Event, error)
	Update(ctx context.Context, event *model.Event) error
	Delete(ctx context.Context, eventID, familyID string) error
	CreateException(ctx context.Context, event *model.Event) error
	CancelOccurrence(ctx context.Context, parentID, familyID string, date time.Time) error
}
