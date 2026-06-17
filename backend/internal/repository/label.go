package repository

import (
	"context"

	"github.com/pidanou/family-board/internal/model"
)

type LabelRepository interface {
	Create(ctx context.Context, label *model.Label) error
	ListByFamilyID(ctx context.Context, familyID string) ([]*model.Label, error)
	Delete(ctx context.Context, labelID, familyID string) error
}
