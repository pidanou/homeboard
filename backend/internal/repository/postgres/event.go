package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pidanou/family-board/internal/model"
)

type EventRepository struct {
	pool *pgxpool.Pool
}

func NewEventRepository(pool *pgxpool.Pool) *EventRepository {
	return &EventRepository{pool: pool}
}

func (r *EventRepository) Create(ctx context.Context, event *model.Event) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO events (id, family_id, title, description, start_at, end_at, all_day, created_by, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		event.ID, event.FamilyID, event.Title, event.Description,
		event.StartAt, event.EndAt, event.AllDay, event.CreatedBy, event.CreatedAt, event.UpdatedAt,
	)
	return err
}

func (r *EventRepository) ListByFamilyAndRange(ctx context.Context, familyID string, from, to time.Time) ([]*model.Event, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, family_id, title, description, start_at, end_at, all_day, created_by, created_at, updated_at
		 FROM events WHERE family_id = $1 AND start_at < $3 AND end_at > $2
		 ORDER BY start_at`,
		familyID, from, to,
	)
	if err != nil {
		return nil, fmt.Errorf("list events: %w", err)
	}
	defer rows.Close()

	events := make([]*model.Event, 0)
	for rows.Next() {
		e := &model.Event{}
		if err := rows.Scan(&e.ID, &e.FamilyID, &e.Title, &e.Description,
			&e.StartAt, &e.EndAt, &e.AllDay, &e.CreatedBy, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, rows.Err()
}

func (r *EventRepository) Update(ctx context.Context, event *model.Event) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE events SET title = $1, description = $2, start_at = $3, end_at = $4, all_day = $5, updated_at = $6
		 WHERE id = $7 AND family_id = $8`,
		event.Title, event.Description, event.StartAt, event.EndAt, event.AllDay, event.UpdatedAt, event.ID, event.FamilyID,
	)
	return err
}

func (r *EventRepository) Delete(ctx context.Context, eventID, familyID string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM events WHERE id = $1 AND family_id = $2`, eventID, familyID)
	return err
}
