package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/google/uuid"
	"github.com/pidanou/homeboard/internal/config"
	"github.com/pidanou/homeboard/internal/model"
	"github.com/pidanou/homeboard/internal/repository"
	"golang.org/x/oauth2"
)

// ErrOIDCEmailNotVerified is returned when the identity provider does not
// assert a verified email — we never auto-link or create an account in that case.
var ErrOIDCEmailNotVerified = errors.New("oidc: provider did not assert a verified email")

// OIDCClaims is a plain projection of the fields we need from an ID token,
// decoupled from *oidc.IDToken so FindOrCreateUser is testable without JWKS/crypto.
type OIDCClaims struct {
	Issuer        string
	Subject       string
	Email         string
	EmailVerified bool
	Name          string
}

type OIDCService struct {
	Provider     *oidc.Provider
	Verifier     *oidc.IDTokenVerifier
	OAuth2Config oauth2.Config
	providerName string

	users            repository.UserRepository
	identities       repository.OIDCIdentityRepository
	issueToken       func(userID string) (string, error)
	registrationGate func(ctx context.Context) error
}

// NewOIDCService performs OIDC discovery against cfg.IssuerURL. Callers should
// treat a returned error as "OIDC unavailable" and start the app with OIDC
// disabled rather than failing the whole process — the IdP may just be briefly
// unreachable.
func NewOIDCService(
	ctx context.Context,
	cfg config.OIDCConfig,
	users repository.UserRepository,
	identities repository.OIDCIdentityRepository,
	issueToken func(userID string) (string, error),
	registrationGate func(ctx context.Context) error,
) (*OIDCService, error) {
	provider, err := oidc.NewProvider(ctx, cfg.IssuerURL)
	if err != nil {
		return nil, fmt.Errorf("oidc discovery: %w", err)
	}

	oauth2Config := oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		RedirectURL:  cfg.RedirectURL,
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "email", "profile"},
	}

	return &OIDCService{
		Provider:         provider,
		Verifier:         provider.Verifier(&oidc.Config{ClientID: cfg.ClientID}),
		OAuth2Config:     oauth2Config,
		providerName:     cfg.ProviderName,
		users:            users,
		identities:       identities,
		issueToken:       issueToken,
		registrationGate: registrationGate,
	}, nil
}

func (s *OIDCService) ProviderName() string { return s.providerName }

// AuthCodeURL builds the redirect URL to the IdP for a fresh login attempt.
func (s *OIDCService) AuthCodeURL(state, nonce, codeVerifier string) string {
	return s.OAuth2Config.AuthCodeURL(state,
		oidc.Nonce(nonce),
		oauth2.S256ChallengeOption(codeVerifier),
	)
}

// Exchange trades an authorization code for tokens and verifies+parses the ID token.
// Callers must separately compare the returned IDToken's Nonce against the one
// they generated at AuthCodeURL time — this package does not do that for you.
func (s *OIDCService) Exchange(ctx context.Context, code, codeVerifier string) (*oidc.IDToken, error) {
	oauth2Token, err := s.OAuth2Config.Exchange(ctx, code, oauth2.VerifierOption(codeVerifier))
	if err != nil {
		return nil, fmt.Errorf("exchange code: %w", err)
	}
	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		return nil, errors.New("no id_token in token response")
	}
	idToken, err := s.Verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return nil, fmt.Errorf("verify id_token: %w", err)
	}
	return idToken, nil
}

// FindOrCreateUser implements the auto-link-by-verified-email policy:
//   - a returning identity (issuer, subject) resolves directly to its user
//   - a new identity with a verified email matching an existing user auto-links
//   - a new identity with a verified email and no match creates a new user,
//     subject to the same registration gate password registration uses
//   - an unverified/absent email is rejected outright — no account is
//     created or linked
func (s *OIDCService) FindOrCreateUser(ctx context.Context, claims OIDCClaims) (*model.User, error) {
	if identity, err := s.identities.GetByIssuerSubject(ctx, claims.Issuer, claims.Subject); err == nil {
		return s.users.GetByID(ctx, identity.UserID)
	}

	if !claims.EmailVerified || claims.Email == "" {
		return nil, ErrOIDCEmailNotVerified
	}

	now := time.Now().UTC()
	email := claims.Email

	if existing, err := s.users.GetByEmail(ctx, email); err == nil {
		identity := &model.OIDCIdentity{
			ID:            uuid.NewString(),
			UserID:        existing.ID,
			Issuer:        claims.Issuer,
			Subject:       claims.Subject,
			Email:         &email,
			EmailVerified: true,
			CreatedAt:     now,
		}
		if err := s.identities.Create(ctx, identity); err != nil {
			return nil, fmt.Errorf("link oidc identity: %w", err)
		}
		return existing, nil
	}

	if err := s.registrationGate(ctx); err != nil {
		return nil, err
	}

	name := claims.Name
	if name == "" {
		name = email
	}
	user := &model.User{
		ID:        uuid.NewString(),
		Email:     email,
		Name:      name,
		Locale:    "en",
		CreatedAt: now,
	}
	identity := &model.OIDCIdentity{
		ID:            uuid.NewString(),
		UserID:        user.ID,
		Issuer:        claims.Issuer,
		Subject:       claims.Subject,
		Email:         &email,
		EmailVerified: true,
		CreatedAt:     now,
	}
	if err := s.identities.CreateWithUser(ctx, user, identity); err != nil {
		return nil, fmt.Errorf("create user from oidc: %w", err)
	}
	return user, nil
}

// IssueToken mints the app's own JWT for a user — reuses AuthService's issuance
// so there is one source of truth for token shape/expiry regardless of login method.
func (s *OIDCService) IssueToken(userID string) (string, error) {
	return s.issueToken(userID)
}
