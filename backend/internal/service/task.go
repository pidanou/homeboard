package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/pidanou/family-board/internal/model"
	"github.com/pidanou/family-board/internal/repository"
)

type TaskService struct {
	tasks repository.TaskRepository
}

func NewTaskService(tasks repository.TaskRepository) *TaskService {
	return &TaskService{tasks: tasks}
}

func (s *TaskService) Create(ctx context.Context, familyID, userID, title, description string, important bool, assignedTo *string, startDate, endDate *time.Time, categoryID *string) (*model.Task, error) {
	now := time.Now().UTC()
	task := &model.Task{
		ID:          uuid.NewString(),
		FamilyID:    familyID,
		Title:       title,
		Description: description,
		Important:   important,
		Status:      "todo",
		AssignedTo:  assignedTo,
		StartDate:   startDate,
		EndDate:     endDate,
		CategoryID:  categoryID,
		CreatedBy:   userID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.tasks.Create(ctx, task); err != nil {
		return nil, err
	}
	return task, nil
}

func (s *TaskService) GetByID(ctx context.Context, taskID, familyID string) (*model.Task, error) {
	return s.tasks.GetByID(ctx, taskID, familyID)
}

func (s *TaskService) ListForFamily(ctx context.Context, familyID string) ([]*model.Task, error) {
	return s.tasks.ListByFamilyID(ctx, familyID)
}

func (s *TaskService) Update(ctx context.Context, task *model.Task) error {
	task.UpdatedAt = time.Now().UTC()
	return s.tasks.Update(ctx, task)
}

func (s *TaskService) Delete(ctx context.Context, taskID, familyID string) error {
	return s.tasks.Delete(ctx, taskID, familyID)
}

func (s *TaskService) Reorder(ctx context.Context, familyID string, ids []string) error {
	return s.tasks.Reorder(ctx, familyID, ids)
}
