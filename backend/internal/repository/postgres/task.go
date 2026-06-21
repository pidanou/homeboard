package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pidanou/homeboard/internal/model"
)

type TaskRepository struct {
	pool *pgxpool.Pool
}

func NewTaskRepository(pool *pgxpool.Pool) *TaskRepository {
	return &TaskRepository{pool: pool}
}

func (r *TaskRepository) Create(ctx context.Context, task *model.Task) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO tasks (id, family_id, title, description, important, status, assigned_to, start_date, end_date, category_id, icon, created_by, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`,
		task.ID, task.FamilyID, task.Title, task.Description, task.Important, task.Status, task.AssignedTo,
		task.StartDate, task.EndDate, task.CategoryID, task.Icon, task.CreatedBy, task.CreatedAt, task.UpdatedAt,
	)
	return err
}

func (r *TaskRepository) GetByID(ctx context.Context, taskID, familyID string) (*model.Task, error) {
	t := &model.Task{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, family_id, title, COALESCE(description, ''), important, status, assigned_to,
		        start_date, end_date, category_id, icon, manual_order, created_by, created_at, updated_at
		 FROM tasks WHERE id = $1 AND family_id = $2`,
		taskID, familyID,
	).Scan(&t.ID, &t.FamilyID, &t.Title, &t.Description, &t.Important, &t.Status, &t.AssignedTo,
		&t.StartDate, &t.EndDate, &t.CategoryID, &t.Icon, &t.ManualOrder, &t.CreatedBy, &t.CreatedAt, &t.UpdatedAt)
	return t, err
}

func (r *TaskRepository) ListByFamilyID(ctx context.Context, familyID string) ([]*model.Task, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, family_id, title, COALESCE(description, ''), important, status, assigned_to,
		        start_date, end_date, category_id, icon, manual_order, created_by, created_at, updated_at
		 FROM tasks
		 WHERE family_id = $1
		 ORDER BY manual_order ASC NULLS LAST, created_at DESC`,
		familyID,
	)
	if err != nil {
		return nil, fmt.Errorf("list tasks: %w", err)
	}
	defer rows.Close()

	tasks := make([]*model.Task, 0)
	for rows.Next() {
		t := &model.Task{}
		if err := rows.Scan(&t.ID, &t.FamilyID, &t.Title, &t.Description, &t.Important, &t.Status, &t.AssignedTo,
			&t.StartDate, &t.EndDate, &t.CategoryID, &t.Icon, &t.ManualOrder, &t.CreatedBy, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}

func (r *TaskRepository) Update(ctx context.Context, task *model.Task) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE tasks SET title = $1, description = $2, important = $3, status = $4, assigned_to = $5,
		  start_date = $6, end_date = $7, category_id = $8, icon = $9, manual_order = $10, updated_at = $11
		 WHERE id = $12 AND family_id = $13`,
		task.Title, task.Description, task.Important, task.Status, task.AssignedTo,
		task.StartDate, task.EndDate, task.CategoryID, task.Icon, task.ManualOrder, task.UpdatedAt, task.ID, task.FamilyID,
	)
	return err
}

func (r *TaskRepository) Reorder(ctx context.Context, familyID string, ids []string) error {
	batch := &pgx.Batch{}
	for i, id := range ids {
		batch.Queue(`UPDATE tasks SET manual_order = $1 WHERE id = $2 AND family_id = $3`, i, id, familyID)
	}
	return r.pool.SendBatch(ctx, batch).Close()
}

func (r *TaskRepository) Delete(ctx context.Context, taskID, familyID string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM tasks WHERE id = $1 AND family_id = $2`, taskID, familyID)
	return err
}
