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

func (s *FamilyService) GetMembers(ctx context.Context, familyID string) ([]*model.FamilyMember, error) {
	real, err := s.families.GetMembers(ctx, familyID)
	if err != nil {
		return nil, err
	}
	virtual, err := s.families.GetUnlinkedVirtualMembers(ctx, familyID)
	if err != nil {
		return nil, err
	}
	for _, vm := range virtual {
		real = append(real, &model.FamilyMember{
			FamilyID: vm.FamilyID,
			UserID:   vm.ID,
			Name:     vm.Name,
			Virtual:  true,
		})
	}
	return real, nil
}

func (s *FamilyService) CreateVirtualMember(ctx context.Context, familyID, name string) (*model.VirtualMember, error) {
	m := &model.VirtualMember{
		ID:        uuid.NewString(),
		FamilyID:  familyID,
		Name:      name,
		CreatedAt: time.Now().UTC(),
	}
	if err := s.families.CreateVirtualMember(ctx, m); err != nil {
		return nil, fmt.Errorf("create virtual member: %w", err)
	}
	return m, nil
}

func (s *FamilyService) DeleteVirtualMember(ctx context.Context, id, familyID string) error {
	return s.families.DeleteVirtualMember(ctx, id, familyID)
}

func (s *FamilyService) GetUnlinkedVirtualMembers(ctx context.Context, familyID string) ([]*model.VirtualMember, error) {
	return s.families.GetUnlinkedVirtualMembers(ctx, familyID)
}

func (s *FamilyService) LinkVirtualMember(ctx context.Context, virtualID, familyID, userID string) error {
	return s.families.LinkVirtualMember(ctx, virtualID, familyID, userID)
}
