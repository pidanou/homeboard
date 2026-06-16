package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pidanou/family-board/internal/model"
)

type InviteRepository struct {
	pool *pgxpool.Pool
}

func NewInviteRepository(pool *pgxpool.Pool) *InviteRepository {
	return &InviteRepository{pool: pool}
}

func (r *InviteRepository) Create(ctx context.Context, invite *model.Invite) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO invites (token, family_id, created_by, created_at, expires_at)
		 VALUES ($1, $2, $3, $4, $5)`,
		invite.Token, invite.FamilyID, invite.CreatedBy, invite.CreatedAt, invite.ExpiresAt,
	)
	return err
}

func (r *InviteRepository) GetByToken(ctx context.Context, token string) (*model.Invite, error) {
	inv := &model.Invite{}
	err := r.pool.QueryRow(ctx,
		`SELECT i.token, i.family_id, f.name, i.created_by, i.created_at, i.expires_at, i.used_at
		 FROM invites i JOIN families f ON f.id = i.family_id
		 WHERE i.token = $1`,
		token,
	).Scan(&inv.Token, &inv.FamilyID, &inv.FamilyName, &inv.CreatedBy, &inv.CreatedAt, &inv.ExpiresAt, &inv.UsedAt)
	if err != nil {
		return nil, fmt.Errorf("get invite: %w", err)
	}
	return inv, nil
}

func (r *InviteRepository) ListByFamilyID(ctx context.Context, familyID string) ([]*model.Invite, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT token, family_id, created_by, created_at, expires_at, used_at
		 FROM invites WHERE family_id = $1 AND used_at IS NULL AND expires_at > NOW()
		 ORDER BY created_at DESC`,
		familyID,
	)
	if err != nil {
		return nil, fmt.Errorf("list invites: %w", err)
	}
	defer rows.Close()

	invites := make([]*model.Invite, 0)
	for rows.Next() {
		inv := &model.Invite{}
		if err := rows.Scan(&inv.Token, &inv.FamilyID, &inv.CreatedBy, &inv.CreatedAt, &inv.ExpiresAt, &inv.UsedAt); err != nil {
			return nil, err
		}
		invites = append(invites, inv)
	}
	return invites, rows.Err()
}

func (r *InviteRepository) Delete(ctx context.Context, token string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM invites WHERE token = $1`, token)
	return err
}

func (r *InviteRepository) MarkUsed(ctx context.Context, token string) error {
	now := time.Now().UTC()
	_, err := r.pool.Exec(ctx,
		`UPDATE invites SET used_at = $1 WHERE token = $2`,
		now, token,
	)
	return err
}
