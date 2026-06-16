package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pidanou/family-board/internal/model"
)

type TaskRepository struct {
	pool *pgxpool.Pool
}

func NewTaskRepository(pool *pgxpool.Pool) *TaskRepository {
	return &TaskRepository{pool: pool}
}

func (r *TaskRepository) Create(ctx context.Context, task *model.Task) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO tasks (id, family_id, title, status, assigned_to, start_date, end_date, created_by, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		task.ID, task.FamilyID, task.Title, task.Status, task.AssignedTo,
		task.StartDate, task.EndDate, task.CreatedBy, task.CreatedAt, task.UpdatedAt,
	)
	return err
}

func (r *TaskRepository) ListByFamilyID(ctx context.Context, familyID string) ([]*model.Task, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, family_id, title, status, assigned_to, start_date, end_date, created_by, created_at, updated_at
		 FROM tasks WHERE family_id = $1 ORDER BY created_at DESC`,
		familyID,
	)
	if err != nil {
		return nil, fmt.Errorf("list tasks: %w", err)
	}
	defer rows.Close()

	tasks := make([]*model.Task, 0)
	for rows.Next() {
		t := &model.Task{}
		if err := rows.Scan(&t.ID, &t.FamilyID, &t.Title, &t.Status, &t.AssignedTo,
			&t.StartDate, &t.EndDate, &t.CreatedBy, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}

func (r *TaskRepository) Update(ctx context.Context, task *model.Task) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE tasks SET title = $1, status = $2, assigned_to = $3, start_date = $4, end_date = $5, updated_at = $6
		 WHERE id = $7 AND family_id = $8`,
		task.Title, task.Status, task.AssignedTo, task.StartDate, task.EndDate, task.UpdatedAt, task.ID, task.FamilyID,
	)
	return err
}

func (r *TaskRepository) Delete(ctx context.Context, taskID, familyID string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM tasks WHERE id = $1 AND family_id = $2`, taskID, familyID)
	return err
}
