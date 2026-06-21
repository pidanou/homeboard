package postgres

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	rrule "github.com/teambition/rrule-go"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pidanou/homeboard/internal/model"
)

type EventRepository struct {
	pool *pgxpool.Pool
}

func NewEventRepository(pool *pgxpool.Pool) *EventRepository {
	return &EventRepository{pool: pool}
}

func (r *EventRepository) Create(ctx context.Context, event *model.Event) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO events (id, family_id, title, description, location, start_at, end_at, all_day,
		  category_id, recurrence_rule, recurrence_parent_id, recurrence_date, cancelled, created_by, created_at, updated_at, type, icon)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18)`,
		event.ID, event.FamilyID, event.Title, event.Description, event.Location,
		event.StartAt, event.EndAt, event.AllDay, event.CategoryID,
		event.RecurrenceRule, event.RecurrenceParentID, event.RecurrenceDate, event.Cancelled,
		event.CreatedBy, event.CreatedAt, event.UpdatedAt, event.Type, event.Icon,
	)
	if err != nil {
		return err
	}
	return r.syncAttendees(ctx, event.ID, event.AttendeeIDs)
}

func (r *EventRepository) ListByFamilyAndRange(ctx context.Context, familyID string, from, to time.Time) ([]*model.Event, error) {
	// Fetch non-exception events (parent rows).
	// Non-recurring: must overlap [from,to).
	// Recurring: must start before `to` (occurrences in range are computed below).
	rows, err := r.pool.Query(ctx,
		`SELECT e.id, e.family_id, e.title, COALESCE(e.description,''), COALESCE(e.location,''),
		        e.start_at, e.end_at, e.all_day, e.category_id,
		        e.recurrence_rule, e.recurrence_parent_id, e.recurrence_date, e.cancelled,
		        e.created_by, e.created_at, e.updated_at, e.type, e.icon,
		        COALESCE(array_agg(DISTINCT ea.user_id) FILTER (WHERE ea.user_id IS NOT NULL), ARRAY[]::text[])
		 FROM events e
		 LEFT JOIN event_attendees ea ON ea.event_id = e.id
		 WHERE e.family_id = $1
		   AND e.recurrence_parent_id IS NULL
		   AND (
		         (e.recurrence_rule IS NULL AND e.start_at < $3 AND e.end_at >= $2)
		      OR (e.recurrence_rule IS NOT NULL AND e.start_at < $3)
		   )
		 GROUP BY e.id
		 ORDER BY e.start_at`,
		familyID, from, to,
	)
	if err != nil {
		return nil, fmt.Errorf("list events: %w", err)
	}
	defer rows.Close()

	var parents []*model.Event
	for rows.Next() {
		e := &model.Event{}
		if err := scanEvent(rows, e); err != nil {
			return nil, err
		}
		parents = append(parents, e)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Fetch all exceptions in range.
	excRows, err := r.pool.Query(ctx,
		`SELECT e.id, e.family_id, e.title, COALESCE(e.description,''), COALESCE(e.location,''),
		        e.start_at, e.end_at, e.all_day, e.category_id,
		        e.recurrence_rule, e.recurrence_parent_id, e.recurrence_date, e.cancelled,
		        e.created_by, e.created_at, e.updated_at, e.type, e.icon,
		        COALESCE(array_agg(DISTINCT ea.user_id) FILTER (WHERE ea.user_id IS NOT NULL), ARRAY[]::text[])
		 FROM events e
		 LEFT JOIN event_attendees ea ON ea.event_id = e.id
		 WHERE e.family_id = $1
		   AND e.recurrence_parent_id IS NOT NULL
		   AND e.recurrence_date >= $2 AND e.recurrence_date < $3
		 GROUP BY e.id`,
		familyID, from, to,
	)
	if err != nil {
		return nil, fmt.Errorf("list exceptions: %w", err)
	}
	defer excRows.Close()

	// exceptions keyed by "parentID::YYYYMMDD"
	exceptions := map[string]*model.Event{}
	for excRows.Next() {
		e := &model.Event{}
		if err := scanEvent(excRows, e); err != nil {
			return nil, err
		}
		if e.RecurrenceParentID != nil && e.RecurrenceDate != nil {
			key := *e.RecurrenceParentID + "::" + e.RecurrenceDate.Format("20060102")
			exceptions[key] = e
		}
	}
	if err := excRows.Err(); err != nil {
		return nil, err
	}

	var result []*model.Event
	for _, e := range parents {
		if e.RecurrenceRule == nil {
			result = append(result, e)
			continue
		}
		// Expand recurring event into occurrences within [from, to).
		occurrences := expandRecurrences(e, from, to)
		for _, occ := range occurrences {
			key := e.ID + "::" + occ.StartAt.Format("20060102")
			if exc, ok := exceptions[key]; ok {
				if exc.Cancelled {
					continue // deleted occurrence
				}
				exc.IsRecurring = true
				result = append(result, exc)
			} else {
				cp := *occ
				cp.IsRecurring = true
				result = append(result, &cp)
			}
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].StartAt.Before(result[j].StartAt)
	})
	return result, nil
}

func expandRecurrences(e *model.Event, from, to time.Time) []*model.Event {
	ruleStr := strings.TrimPrefix(*e.RecurrenceRule, "RRULE:")
	rOption, err := rrule.StrToROption(ruleStr)
	if err != nil {
		return nil
	}
	rOption.Dtstart = e.StartAt
	rule, err := rrule.NewRRule(*rOption)
	if err != nil {
		return nil
	}

	duration := e.EndAt.Sub(e.StartAt)
	var result []*model.Event
	for _, occ := range rule.Between(from, to, true) {
		cp := *e
		cp.StartAt = occ
		cp.EndAt = occ.Add(duration)
		occDate := occ
		cp.RecurrenceDate = &occDate
		cp.ID = fmt.Sprintf("%s::%s", e.ID, occ.Format("20060102"))
		result = append(result, &cp)
	}
	return result
}

// CreateException inserts a modified occurrence row (exception).
func (r *EventRepository) CreateException(ctx context.Context, event *model.Event) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO events (id, family_id, title, description, location, start_at, end_at, all_day,
		  category_id, recurrence_rule, recurrence_parent_id, recurrence_date, cancelled, created_by, created_at, updated_at, type, icon)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18)`,
		event.ID, event.FamilyID, event.Title, event.Description, event.Location,
		event.StartAt, event.EndAt, event.AllDay, event.CategoryID,
		nil, event.RecurrenceParentID, event.RecurrenceDate, false,
		event.CreatedBy, event.CreatedAt, event.UpdatedAt, "default", nil,
	)
	if err != nil {
		return err
	}
	return r.syncAttendees(ctx, event.ID, event.AttendeeIDs)
}

// CancelOccurrence marks a single occurrence as cancelled by inserting a cancelled exception row.
func (r *EventRepository) CancelOccurrence(ctx context.Context, parentID, familyID string, date time.Time) error {
	// Fetch parent to get created_by.
	var createdBy string
	if err := r.pool.QueryRow(ctx,
		`SELECT created_by FROM events WHERE id = $1 AND family_id = $2`, parentID, familyID,
	).Scan(&createdBy); err != nil {
		return fmt.Errorf("parent event not found: %w", err)
	}
	now := time.Now().UTC()
	id := fmt.Sprintf("%s::%s", parentID, date.Format("20060102"))
	_, err := r.pool.Exec(ctx,
		`INSERT INTO events (id, family_id, title, description, location, start_at, end_at, all_day,
		  recurrence_parent_id, recurrence_date, cancelled, created_by, created_at, updated_at)
		 VALUES ($1,$2,'',' ',' ',$3,$3,false,$4,$5,true,$6,$7,$7)
		 ON CONFLICT (id) DO UPDATE SET cancelled = true`,
		id, familyID, date, parentID, date, createdBy, now,
	)
	return err
}

func (r *EventRepository) Update(ctx context.Context, event *model.Event) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE events SET title=$1, description=$2, location=$3, start_at=$4, end_at=$5,
		  all_day=$6, category_id=$7, recurrence_rule=$8, icon=$9, updated_at=$10
		 WHERE id=$11 AND family_id=$12`,
		event.Title, event.Description, event.Location, event.StartAt, event.EndAt,
		event.AllDay, event.CategoryID, event.RecurrenceRule, event.Icon, event.UpdatedAt,
		event.ID, event.FamilyID,
	)
	if err != nil {
		return err
	}
	return r.syncAttendees(ctx, event.ID, event.AttendeeIDs)
}

func (r *EventRepository) Delete(ctx context.Context, eventID, familyID string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM events WHERE id=$1 AND family_id=$2`, eventID, familyID)
	return err
}

func (r *EventRepository) syncAttendees(ctx context.Context, eventID string, userIDs []string) error {
	if _, err := r.pool.Exec(ctx, `DELETE FROM event_attendees WHERE event_id=$1`, eventID); err != nil {
		return err
	}
	for _, uid := range userIDs {
		if _, err := r.pool.Exec(ctx, `INSERT INTO event_attendees (event_id, user_id) VALUES ($1,$2)`, eventID, uid); err != nil {
			return err
		}
	}
	return nil
}

type scannable interface {
	Scan(dest ...any) error
}

func scanEvent(row scannable, e *model.Event) error {
	return row.Scan(
		&e.ID, &e.FamilyID, &e.Title, &e.Description, &e.Location,
		&e.StartAt, &e.EndAt, &e.AllDay, &e.CategoryID,
		&e.RecurrenceRule, &e.RecurrenceParentID, &e.RecurrenceDate, &e.Cancelled,
		&e.CreatedBy, &e.CreatedAt, &e.UpdatedAt, &e.Type, &e.Icon,
		&e.AttendeeIDs,
	)
}
