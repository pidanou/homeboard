package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/pidanou/family-board/internal/model"
	"github.com/pidanou/family-board/internal/repository"
)

type LabelService struct {
	labels repository.LabelRepository
}

func NewLabelService(labels repository.LabelRepository) *LabelService {
	return &LabelService{labels: labels}
}

func (s *LabelService) Create(ctx context.Context, familyID, name, color string) (*model.Label, error) {
	label := &model.Label{
		ID:        uuid.NewString(),
		FamilyID:  familyID,
		Name:      name,
		Color:     color,
		CreatedAt: time.Now().UTC(),
	}
	if err := s.labels.Create(ctx, label); err != nil {
		return nil, err
	}
	return label, nil
}

func (s *LabelService) ListForFamily(ctx context.Context, familyID string) ([]*model.Label, error) {
	return s.labels.ListByFamilyID(ctx, familyID)
}

func (s *LabelService) Delete(ctx context.Context, labelID, familyID string) error {
	return s.labels.Delete(ctx, labelID, familyID)
}
