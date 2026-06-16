package repository

import (
	"context"

	"github.com/pidanou/family-board/internal/model"
)

type FamilyRepository interface {
	Create(ctx context.Context, family *model.Family) error
	GetByID(ctx context.Context, id string) (*model.Family, error)
	AddMember(ctx context.Context, member *model.FamilyMember) error
	GetMembers(ctx context.Context, familyID string) ([]*model.FamilyMember, error)
	GetFamiliesByUserID(ctx context.Context, userID string) ([]*model.Family, error)
}
