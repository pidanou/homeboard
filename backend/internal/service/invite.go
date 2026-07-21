package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/pidanou/homeboard/internal/model"
	"github.com/pidanou/homeboard/internal/repository"
)

var (
	ErrInviteNotFound = errors.New("invite not found")
	ErrInviteNoEmail  = errors.New("invite has no email address on file")
	ErrInviteUsed     = errors.New("invite already used")
	ErrInviteExpired  = errors.New("invite expired")
)

type InviteService struct {
	invites    repository.InviteRepository
	families   repository.HouseholdRepository
	mailer     *EmailService
	appBaseURL string
}

func NewInviteService(invites repository.InviteRepository, families repository.HouseholdRepository, mailer *EmailService, appBaseURL string) *InviteService {
	return &InviteService{invites: invites, families: families, mailer: mailer, appBaseURL: strings.TrimRight(appBaseURL, "/")}
}

// Create generates a new invite link for the family, revoking any previous one.
// If email is non-empty, the invite link is also emailed to that address.
func (s *InviteService) Create(ctx context.Context, familyID, createdBy, email string) (*model.Invite, error) {
	if err := s.invites.DeleteByFamilyID(ctx, familyID); err != nil {
		return nil, fmt.Errorf("clear existing invites: %w", err)
	}
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
	if email != "" {
		invite.Email = &email
	}

	if err := s.invites.Create(ctx, invite); err != nil {
		return nil, fmt.Errorf("create invite: %w", err)
	}

	if email != "" {
		s.sendInviteEmail(ctx, invite)
	}
	return invite, nil
}

// Resend re-emails an existing, still-valid invite to the address it was created with.
func (s *InviteService) Resend(ctx context.Context, token string) error {
	invite, err := s.invites.GetByToken(ctx, token)
	if err != nil {
		return ErrInviteNotFound
	}
	if invite.Email == nil {
		return ErrInviteNoEmail
	}
	if invite.UsedAt != nil {
		return ErrInviteUsed
	}
	if time.Now().UTC().After(invite.ExpiresAt) {
		return ErrInviteExpired
	}
	s.sendInviteEmail(ctx, invite)
	return nil
}

func (s *InviteService) sendInviteEmail(ctx context.Context, invite *model.Invite) {
	familyName := invite.FamilyName
	if familyName == "" {
		if family, err := s.families.GetByID(ctx, invite.FamilyID); err == nil {
			familyName = family.Name
		}
	}
	s.mailer.Send(*invite.Email, emailData{
		Subject:    "You've been invited to Family Board",
		Heading:    "You're invited!",
		Name:       *invite.Email,
		Body:       fmt.Sprintf("You've been invited to join %s on Family Board.", familyName),
		ButtonText: "Accept invite",
		ButtonURL:  fmt.Sprintf("%s/invite/%s", s.appBaseURL, invite.Token),
	})
}

type AcceptResult struct {
	FamilyID               string                 `json:"family_id"`
	UnlinkedVirtualMembers []*model.VirtualMember `json:"unlinked_virtual_members"`
}

func (s *InviteService) Accept(ctx context.Context, token, userID string) (*AcceptResult, error) {
	invite, err := s.invites.GetByToken(ctx, token)
	if err != nil {
		return nil, ErrInviteNotFound
	}
	if invite.UsedAt != nil {
		return nil, ErrInviteUsed
	}
	if time.Now().UTC().After(invite.ExpiresAt) {
		return nil, ErrInviteExpired
	}

	member := &model.HouseholdMember{
		FamilyID: invite.FamilyID,
		UserID:   userID,
		Role:     "member",
		JoinedAt: time.Now().UTC(),
	}
	if err := s.families.AddMember(ctx, member); err != nil {
		return nil, fmt.Errorf("add member: %w", err)
	}

	if err := s.invites.MarkUsed(ctx, token); err != nil {
		return nil, err
	}

	unlinked, err := s.families.GetUnlinkedVirtualMembers(ctx, invite.FamilyID)
	if err != nil {
		unlinked = nil // non-fatal
	}

	return &AcceptResult{
		FamilyID:               invite.FamilyID,
		UnlinkedVirtualMembers: unlinked,
	}, nil
}

func (s *InviteService) GetByToken(ctx context.Context, token string) (*model.Invite, error) {
	return s.invites.GetByToken(ctx, token)
}

func (s *InviteService) Validate(ctx context.Context, token string) error {
	invite, err := s.invites.GetByToken(ctx, token)
	if err != nil {
		return ErrInviteNotFound
	}
	if invite.UsedAt != nil {
		return ErrInviteUsed
	}
	if time.Now().UTC().After(invite.ExpiresAt) {
		return ErrInviteExpired
	}
	return nil
}

func (s *InviteService) ListForFamily(ctx context.Context, familyID string) ([]*model.Invite, error) {
	return s.invites.ListByFamilyID(ctx, familyID)
}

func (s *InviteService) Delete(ctx context.Context, token string) error {
	return s.invites.Delete(ctx, token)
}
