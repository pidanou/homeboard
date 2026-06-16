package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/pidanou/family-board/internal/model"
	"github.com/pidanou/family-board/internal/repository"
)

type InviteService struct {
	invites  repository.InviteRepository
	families repository.FamilyRepository
}

func NewInviteService(invites repository.InviteRepository, families repository.FamilyRepository) *InviteService {
	return &InviteService{invites: invites, families: families}
}

func (s *InviteService) Create(ctx context.Context, familyID, createdBy string) (*model.Invite, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	invite := &model.Invite{
		Token:     hex.EncodeToString(b),
		FamilyID:  familyID,
		CreatedBy: createdBy,
		CreatedAt: time.Now().UTC(),
		ExpiresAt: time.Now().UTC().Add(7 * 24 * time.Hour),
	}

	if err := s.invites.Create(ctx, invite); err != nil {
		return nil, fmt.Errorf("create invite: %w", err)
	}
	return invite, nil
}

func (s *InviteService) Accept(ctx context.Context, token, userID string) error {
	invite, err := s.invites.GetByToken(ctx, token)
	if err != nil {
		return errors.New("invite not found")
	}
	if invite.UsedAt != nil {
		return errors.New("invite already used")
	}
	if time.Now().UTC().After(invite.ExpiresAt) {
		return errors.New("invite expired")
	}

	member := &model.FamilyMember{
		FamilyID: invite.FamilyID,
		UserID:   userID,
		Role:     "member",
		JoinedAt: time.Now().UTC(),
	}
	if err := s.families.AddMember(ctx, member); err != nil {
		return fmt.Errorf("add member: %w", err)
	}

	return s.invites.MarkUsed(ctx, token)
}

func (s *InviteService) GetByToken(ctx context.Context, token string) (*model.Invite, error) {
	return s.invites.GetByToken(ctx, token)
}

func (s *InviteService) ListForFamily(ctx context.Context, familyID string) ([]*model.Invite, error) {
	return s.invites.ListByFamilyID(ctx, familyID)
}

func (s *InviteService) Delete(ctx context.Context, token string) error {
	return s.invites.Delete(ctx, token)
}
