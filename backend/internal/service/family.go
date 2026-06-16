package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/pidanou/family-board/internal/model"
	"github.com/pidanou/family-board/internal/repository"
)

type FamilyService struct {
	families repository.FamilyRepository
}

func NewFamilyService(families repository.FamilyRepository) *FamilyService {
	return &FamilyService{families: families}
}

func (s *FamilyService) Create(ctx context.Context, name, ownerID string) (*model.Family, error) {
	family := &model.Family{
		ID:        uuid.NewString(),
		Name:      name,
		CreatedAt: time.Now().UTC(),
	}

	if err := s.families.Create(ctx, family); err != nil {
		return nil, fmt.Errorf("create family: %w", err)
	}

	member := &model.FamilyMember{
		FamilyID: family.ID,
		UserID:   ownerID,
		Role:     "admin",
		JoinedAt: time.Now().UTC(),
	}
	if err := s.families.AddMember(ctx, member); err != nil {
		return nil, fmt.Errorf("add owner as member: %w", err)
	}

	return family, nil
}

func (s *FamilyService) GetByID(ctx context.Context, id string) (*model.Family, error) {
	return s.families.GetByID(ctx, id)
}

func (s *FamilyService) ListForUser(ctx context.Context, userID string) ([]*model.Family, error) {
	return s.families.GetFamiliesByUserID(ctx, userID)
}
