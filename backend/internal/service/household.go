package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/pidanou/family-board/internal/model"
	"github.com/pidanou/family-board/internal/repository"
)

type HouseholdService struct {
	families repository.HouseholdRepository
}

func NewHouseholdService(families repository.HouseholdRepository) *HouseholdService {
	return &HouseholdService{families: families}
}

func (s *HouseholdService) Create(ctx context.Context, name, ownerID string) (*model.Household, error) {
	family := &model.Household{
		ID:        uuid.NewString(),
		Name:      name,
		CreatedAt: time.Now().UTC(),
	}

	if err := s.families.Create(ctx, family); err != nil {
		return nil, fmt.Errorf("create family: %w", err)
	}

	member := &model.HouseholdMember{
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

func (s *HouseholdService) GetByID(ctx context.Context, id string) (*model.Household, error) {
	return s.families.GetByID(ctx, id)
}

func (s *HouseholdService) ListForUser(ctx context.Context, userID string) ([]*model.Household, error) {
	return s.families.GetHouseholdsByUserID(ctx, userID)
}

func (s *HouseholdService) GetMembers(ctx context.Context, familyID string) ([]*model.HouseholdMember, error) {
	real, err := s.families.GetMembers(ctx, familyID)
	if err != nil {
		return nil, err
	}
	virtual, err := s.families.GetUnlinkedVirtualMembers(ctx, familyID)
	if err != nil {
		return nil, err
	}
	for _, vm := range virtual {
		real = append(real, &model.HouseholdMember{
			FamilyID: vm.FamilyID,
			UserID:   vm.ID,
			Name:     vm.Name,
			Virtual:  true,
		})
	}
	return real, nil
}

func (s *HouseholdService) CreateVirtualMember(ctx context.Context, familyID, name, callerID string) (*model.VirtualMember, error) {
	role, err := s.families.GetMemberRole(ctx, callerID, familyID)
	if err != nil || role != "admin" {
		return nil, fmt.Errorf("only admins can create virtual members")
	}
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

func (s *HouseholdService) DeleteVirtualMember(ctx context.Context, id, familyID, callerID string) error {
	role, err := s.families.GetMemberRole(ctx, callerID, familyID)
	if err != nil || role != "admin" {
		return fmt.Errorf("only admins can remove virtual members")
	}
	return s.families.DeleteVirtualMember(ctx, id, familyID)
}

func (s *HouseholdService) GetUnlinkedVirtualMembers(ctx context.Context, familyID string) ([]*model.VirtualMember, error) {
	return s.families.GetUnlinkedVirtualMembers(ctx, familyID)
}

func (s *HouseholdService) LinkVirtualMember(ctx context.Context, virtualID, familyID, userID string) error {
	return s.families.LinkVirtualMember(ctx, virtualID, familyID, userID)
}

func (s *HouseholdService) GetMemberRole(ctx context.Context, userID, familyID string) (string, error) {
	return s.families.GetMemberRole(ctx, userID, familyID)
}

func (s *HouseholdService) UpdateMemberRole(ctx context.Context, targetID, familyID, role, callerID string) error {
	if role != "admin" && role != "member" {
		return fmt.Errorf("invalid role")
	}
	callerRole, err := s.families.GetMemberRole(ctx, callerID, familyID)
	if err != nil || callerRole != "admin" {
		return fmt.Errorf("only admins can change roles")
	}
	if callerID == targetID {
		return fmt.Errorf("cannot change your own role")
	}
	if role == "member" {
		count, err := s.families.CountAdmins(ctx, familyID)
		if err != nil {
			return err
		}
		if count <= 1 {
			return fmt.Errorf("cannot demote the last admin")
		}
	}
	return s.families.UpdateMemberRole(ctx, targetID, familyID, role)
}

func (s *HouseholdService) RemoveMember(ctx context.Context, userID, familyID, callerID string) error {
	if userID == callerID {
		return fmt.Errorf("cannot remove yourself")
	}
	callerRole, err := s.families.GetMemberRole(ctx, callerID, familyID)
	if err != nil || callerRole != "admin" {
		return fmt.Errorf("only admins can remove members")
	}
	return s.families.RemoveMember(ctx, userID, familyID)
}
