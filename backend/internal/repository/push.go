package repository

import (
	"context"

	"github.com/pidanou/homeboard/internal/model"
)

type PushRepository interface {
	Save(ctx context.Context, sub *model.PushSubscription) error
	Delete(ctx context.Context, userID, endpoint string) error
	ListForFamily(ctx context.Context, familyID string) ([]*model.PushSubscription, error)
	DeleteByEndpoint(ctx context.Context, endpoint string) error
}
