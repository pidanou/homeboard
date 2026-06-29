package repository

import (
	"context"

	"github.com/pidanou/homeboard/internal/model"
)

type HouseholdRepository interface {
	Create(ctx context.Context, family *model.Household) error
	GetByID(ctx context.Context, id string) (*model.Household, error)
	AddMember(ctx context.Context, member *model.HouseholdMember) error
	GetMembers(ctx context.Context, familyID string) ([]*model.HouseholdMember, error)
	GetHouseholdsByUserID(ctx context.Context, userID string) ([]*model.Household, error)
	CreateVirtualMember(ctx context.Context, m *model.VirtualMember) error
	DeleteVirtualMember(ctx context.Context, id, familyID string) error
	GetUnlinkedVirtualMembers(ctx context.Context, familyID string) ([]*model.VirtualMember, error)
	LinkVirtualMember(ctx context.Context, virtualID, familyID, userID string) error
	RemoveMember(ctx context.Context, userID, familyID string) error
	GetMemberRole(ctx context.Context, userID, familyID string) (string, error)
	UpdateMemberRole(ctx context.Context, userID, familyID, role string) error
	CountAdmins(ctx context.Context, familyID string) (int, error)
	UpdateName(ctx context.Context, id, name string) error
	Exists(ctx context.Context) (bool, error)
}
