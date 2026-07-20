package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pidanou/homeboard/internal/model"
)

type PasswordResetRepository struct {
	pool *pgxpool.Pool
}

func NewPasswordResetRepository(pool *pgxpool.Pool) *PasswordResetRepository {
	return &PasswordResetRepository{pool: pool}
}

func (r *PasswordResetRepository) Create(ctx context.Context, token *model.PasswordResetToken) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO password_reset_tokens (token, user_id, created_at, expires_at)
		 VALUES ($1, $2, $3, $4)`,
		token.Token, token.UserID, token.CreatedAt, token.ExpiresAt,
	)
	return err
}

func (r *PasswordResetRepository) GetByToken(ctx context.Context, token string) (*model.PasswordResetToken, error) {
	t := &model.PasswordResetToken{}
	err := r.pool.QueryRow(ctx,
		`SELECT token, user_id, created_at, expires_at, used_at FROM password_reset_tokens WHERE token = $1`,
		token,
	).Scan(&t.Token, &t.UserID, &t.CreatedAt, &t.ExpiresAt, &t.UsedAt)
	if err != nil {
		return nil, fmt.Errorf("get password reset token: %w", err)
	}
	return t, nil
}

func (r *PasswordResetRepository) MarkUsed(ctx context.Context, token string) error {
	now := time.Now().UTC()
	_, err := r.pool.Exec(ctx, `UPDATE password_reset_tokens SET used_at = $1 WHERE token = $2`, now, token)
	return err
}

func (r *PasswordResetRepository) DeleteByUserID(ctx context.Context, userID string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM password_reset_tokens WHERE user_id = $1`, userID)
	return err
}
