package repository

import (
	"context"

	"github.com/pidanou/homeboard/internal/model"
)

type CalendarExportTokenRepository interface {
	Create(ctx context.Context, token *model.CalendarExportToken) error
	GetByFamilyID(ctx context.Context, familyID string) (*model.CalendarExportToken, error)
	GetByToken(ctx context.Context, token string) (*model.CalendarExportToken, error)
	DeleteByFamilyID(ctx context.Context, familyID string) error
}

type CalendarSubscriptionRepository interface {
	Create(ctx context.Context, sub *model.CalendarSubscription) error
	ListByFamilyID(ctx context.Context, familyID string) ([]*model.CalendarSubscription, error)
	Get(ctx context.Context, id string) (*model.CalendarSubscription, error)
	Delete(ctx context.Context, id, familyID string) error
	ListAll(ctx context.Context) ([]*model.CalendarSubscription, error)
	UpdateSyncResult(ctx context.Context, id string, syncErr *string) error
}
