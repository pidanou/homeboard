package repository

import (
	"context"

	"github.com/pidanou/homeboard/internal/model"
)

type CategoryRepository interface {
	Create(ctx context.Context, category *model.Category) error
	ListByFamilyID(ctx context.Context, familyID string) ([]*model.Category, error)
	Delete(ctx context.Context, categoryID, familyID string) error
	Update(ctx context.Context, category *model.Category) error
}
