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
		`INSERT INTO tasks (id, family_id, title, description, priority, status, assigned_to, start_date, end_date, created_by, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
		task.ID, task.FamilyID, task.Title, task.Description, task.Priority, task.Status, task.AssignedTo,
		task.StartDate, task.EndDate, task.CreatedBy, task.CreatedAt, task.UpdatedAt,
	)
	if err != nil {
		return err
	}
	return r.syncLabels(ctx, task.ID, task.LabelIDs)
}

func (r *TaskRepository) ListByFamilyID(ctx context.Context, familyID string) ([]*model.Task, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT t.id, t.family_id, t.title, COALESCE(t.description, ''), t.priority, t.status, t.assigned_to,
		        t.start_date, t.end_date, t.created_by, t.created_at, t.updated_at,
		        COALESCE(array_agg(tl.label_id) FILTER (WHERE tl.label_id IS NOT NULL), ARRAY[]::text[])
		 FROM tasks t
		 LEFT JOIN task_labels tl ON tl.task_id = t.id
		 WHERE t.family_id = $1
		 GROUP BY t.id
		 ORDER BY t.created_at DESC`,
		familyID,
	)
	if err != nil {
		return nil, fmt.Errorf("list tasks: %w", err)
	}
	defer rows.Close()

	tasks := make([]*model.Task, 0)
	for rows.Next() {
		t := &model.Task{}
		if err := rows.Scan(&t.ID, &t.FamilyID, &t.Title, &t.Description, &t.Priority, &t.Status, &t.AssignedTo,
			&t.StartDate, &t.EndDate, &t.CreatedBy, &t.CreatedAt, &t.UpdatedAt, &t.LabelIDs); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}

func (r *TaskRepository) Update(ctx context.Context, task *model.Task) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE tasks SET title = $1, description = $2, priority = $3, status = $4, assigned_to = $5, start_date = $6, end_date = $7, updated_at = $8
		 WHERE id = $9 AND family_id = $10`,
		task.Title, task.Description, task.Priority, task.Status, task.AssignedTo,
		task.StartDate, task.EndDate, task.UpdatedAt, task.ID, task.FamilyID,
	)
	if err != nil {
		return err
	}
	return r.syncLabels(ctx, task.ID, task.LabelIDs)
}

func (r *TaskRepository) Delete(ctx context.Context, taskID, familyID string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM tasks WHERE id = $1 AND family_id = $2`, taskID, familyID)
	return err
}

func (r *TaskRepository) syncLabels(ctx context.Context, taskID string, labelIDs []string) error {
	if _, err := r.pool.Exec(ctx, `DELETE FROM task_labels WHERE task_id = $1`, taskID); err != nil {
		return err
	}
	for _, id := range labelIDs {
		if _, err := r.pool.Exec(ctx, `INSERT INTO task_labels (task_id, label_id) VALUES ($1, $2)`, taskID, id); err != nil {
			return err
		}
	}
	return nil
}
