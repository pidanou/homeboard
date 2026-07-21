package repository

import (
	"context"

	"github.com/pidanou/homeboard/internal/model"
)

type PasswordResetRepository interface {
	Create(ctx context.Context, token *model.PasswordResetToken) error
	GetByToken(ctx context.Context, token string) (*model.PasswordResetToken, error)
	MarkUsed(ctx context.Context, token string) error
	DeleteByUserID(ctx context.Context, userID string) error
}
