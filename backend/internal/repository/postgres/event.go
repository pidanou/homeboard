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
		`INSERT INTO events (id, family_id, title, description, location, start_at, end_at, all_day, created_by, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		event.ID, event.FamilyID, event.Title, event.Description, event.Location,
		event.StartAt, event.EndAt, event.AllDay, event.CreatedBy, event.CreatedAt, event.UpdatedAt,
	)
	if err != nil {
		return err
	}
	if err := r.syncAttendees(ctx, event.ID, event.AttendeeIDs); err != nil {
		return err
	}
	return r.syncLabels(ctx, event.ID, event.LabelIDs)
}

func (r *EventRepository) ListByFamilyAndRange(ctx context.Context, familyID string, from, to time.Time) ([]*model.Event, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT e.id, e.family_id, e.title, COALESCE(e.description, ''), COALESCE(e.location, ''),
		        e.start_at, e.end_at, e.all_day, e.created_by, e.created_at, e.updated_at,
		        COALESCE(array_agg(DISTINCT ea.user_id) FILTER (WHERE ea.user_id IS NOT NULL), ARRAY[]::text[]),
		        COALESCE(array_agg(DISTINCT el.label_id) FILTER (WHERE el.label_id IS NOT NULL), ARRAY[]::text[])
		 FROM events e
		 LEFT JOIN event_attendees ea ON ea.event_id = e.id
		 LEFT JOIN event_labels el ON el.event_id = e.id
		 WHERE e.family_id = $1 AND e.start_at < $3 AND e.end_at > $2
		 GROUP BY e.id
		 ORDER BY e.start_at`,
		familyID, from, to,
	)
	if err != nil {
		return nil, fmt.Errorf("list events: %w", err)
	}
	defer rows.Close()

	events := make([]*model.Event, 0)
	for rows.Next() {
		e := &model.Event{}
		if err := rows.Scan(&e.ID, &e.FamilyID, &e.Title, &e.Description, &e.Location,
			&e.StartAt, &e.EndAt, &e.AllDay, &e.CreatedBy, &e.CreatedAt, &e.UpdatedAt,
			&e.AttendeeIDs, &e.LabelIDs); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, rows.Err()
}

func (r *EventRepository) Update(ctx context.Context, event *model.Event) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE events SET title = $1, description = $2, location = $3, start_at = $4, end_at = $5, all_day = $6, updated_at = $7
		 WHERE id = $8 AND family_id = $9`,
		event.Title, event.Description, event.Location, event.StartAt, event.EndAt, event.AllDay,
		event.UpdatedAt, event.ID, event.FamilyID,
	)
	if err != nil {
		return err
	}
	if err := r.syncAttendees(ctx, event.ID, event.AttendeeIDs); err != nil {
		return err
	}
	return r.syncLabels(ctx, event.ID, event.LabelIDs)
}

func (r *EventRepository) Delete(ctx context.Context, eventID, familyID string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM events WHERE id = $1 AND family_id = $2`, eventID, familyID)
	return err
}

func (r *EventRepository) syncAttendees(ctx context.Context, eventID string, userIDs []string) error {
	if _, err := r.pool.Exec(ctx, `DELETE FROM event_attendees WHERE event_id = $1`, eventID); err != nil {
		return err
	}
	for _, uid := range userIDs {
		if _, err := r.pool.Exec(ctx, `INSERT INTO event_attendees (event_id, user_id) VALUES ($1, $2)`, eventID, uid); err != nil {
			return err
		}
	}
	return nil
}

func (r *EventRepository) syncLabels(ctx context.Context, eventID string, labelIDs []string) error {
	if _, err := r.pool.Exec(ctx, `DELETE FROM event_labels WHERE event_id = $1`, eventID); err != nil {
		return err
	}
	for _, id := range labelIDs {
		if _, err := r.pool.Exec(ctx, `INSERT INTO event_labels (event_id, label_id) VALUES ($1, $2)`, eventID, id); err != nil {
			return err
		}
	}
	return nil
}
