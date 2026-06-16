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

func (s *TaskService) Create(ctx context.Context, familyID, userID, title string, startDate, endDate *time.Time) (*model.Task, error) {
	now := time.Now().UTC()
	task := &model.Task{
		ID:        uuid.NewString(),
		FamilyID:  familyID,
		Title:     title,
		Status:    "todo",
		StartDate: startDate,
		EndDate:   endDate,
		CreatedBy: userID,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.tasks.Create(ctx, task); err != nil {
		return nil, err
	}
	return task, nil
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
