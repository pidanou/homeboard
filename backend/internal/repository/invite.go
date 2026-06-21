package repository

import (
	"context"

	"github.com/pidanou/homeboard/internal/model"
)

type InviteRepository interface {
	Create(ctx context.Context, invite *model.Invite) error
	GetByToken(ctx context.Context, token string) (*model.Invite, error)
	ListByFamilyID(ctx context.Context, familyID string) ([]*model.Invite, error)
	MarkUsed(ctx context.Context, token string) error
	Delete(ctx context.Context, token string) error
	DeleteByFamilyID(ctx context.Context, familyID string) error
}
