package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pidanou/homeboard/internal/model"
)

type OIDCIdentityRepository struct {
	pool *pgxpool.Pool
}

func NewOIDCIdentityRepository(pool *pgxpool.Pool) *OIDCIdentityRepository {
	return &OIDCIdentityRepository{pool: pool}
}

func (r *OIDCIdentityRepository) Create(ctx context.Context, identity *model.OIDCIdentity) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO oidc_identities (id, user_id, issuer, subject, email, email_verified, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		identity.ID, identity.UserID, identity.Issuer, identity.Subject, identity.Email, identity.EmailVerified, identity.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("create oidc identity: %w", err)
	}
	return nil
}

func (r *OIDCIdentityRepository) CreateWithUser(ctx context.Context, user *model.User, identity *model.OIDCIdentity) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	if _, err := tx.Exec(ctx,
		`INSERT INTO users (id, email, password_hash, name, avatar_url, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		user.ID, user.Email, user.PasswordHash, user.Name, user.AvatarURL, user.CreatedAt,
	); err != nil {
		return fmt.Errorf("create user: %w", err)
	}

	if _, err := tx.Exec(ctx,
		`INSERT INTO oidc_identities (id, user_id, issuer, subject, email, email_verified, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		identity.ID, identity.UserID, identity.Issuer, identity.Subject, identity.Email, identity.EmailVerified, identity.CreatedAt,
	); err != nil {
		return fmt.Errorf("create oidc identity: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}

func (r *OIDCIdentityRepository) GetByIssuerSubject(ctx context.Context, issuer, subject string) (*model.OIDCIdentity, error) {
	identity := &model.OIDCIdentity{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, issuer, subject, email, email_verified, created_at
		 FROM oidc_identities WHERE issuer = $1 AND subject = $2`,
		issuer, subject,
	).Scan(&identity.ID, &identity.UserID, &identity.Issuer, &identity.Subject, &identity.Email, &identity.EmailVerified, &identity.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("get oidc identity: %w", err)
	}
	return identity, nil
}
