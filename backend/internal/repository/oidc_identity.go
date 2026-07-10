package repository

import (
	"context"

	"github.com/pidanou/homeboard/internal/model"
)

type OIDCIdentityRepository interface {
	// Create links an OIDC identity to an existing user (auto-link path).
	Create(ctx context.Context, identity *model.OIDCIdentity) error
	// CreateWithUser creates a new user and its OIDC identity atomically (new-account-via-OIDC path).
	CreateWithUser(ctx context.Context, user *model.User, identity *model.OIDCIdentity) error
	GetByIssuerSubject(ctx context.Context, issuer, subject string) (*model.OIDCIdentity, error)
}
