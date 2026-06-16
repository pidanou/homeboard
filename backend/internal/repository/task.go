package repository

import (
	"context"

	"github.com/pidanou/family-board/internal/model"
)

type TaskRepository interface {
	Create(ctx context.Context, task *model.Task) error
	ListByFamilyID(ctx context.Context, familyID string) ([]*model.Task, error)
	Update(ctx context.Context, task *model.Task) error
	Delete(ctx context.Context, taskID, familyID string) error
}
