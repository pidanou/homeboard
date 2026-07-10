package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pidanou/homeboard/internal/model"
)

type CalendarSubscriptionRepository struct {
	pool *pgxpool.Pool
}

func NewCalendarSubscriptionRepository(pool *pgxpool.Pool) *CalendarSubscriptionRepository {
	return &CalendarSubscriptionRepository{pool: pool}
}

func (r *CalendarSubscriptionRepository) Create(ctx context.Context, sub *model.CalendarSubscription) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO calendar_subscriptions (id, family_id, name, url, created_by, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		sub.ID, sub.FamilyID, sub.Name, sub.URL, sub.CreatedBy, sub.CreatedAt,
	)
	return err
}

func (r *CalendarSubscriptionRepository) ListByFamilyID(ctx context.Context, familyID string) ([]*model.CalendarSubscription, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, family_id, name, url, created_by, created_at, last_synced_at, last_sync_error
		 FROM calendar_subscriptions WHERE family_id = $1 ORDER BY created_at`,
		familyID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*model.CalendarSubscription
	for rows.Next() {
		s := &model.CalendarSubscription{}
		if err := rows.Scan(&s.ID, &s.FamilyID, &s.Name, &s.URL, &s.CreatedBy, &s.CreatedAt, &s.LastSyncedAt, &s.LastSyncError); err != nil {
			return nil, err
		}
		result = append(result, s)
	}
	return result, rows.Err()
}

func (r *CalendarSubscriptionRepository) Get(ctx context.Context, id string) (*model.CalendarSubscription, error) {
	s := &model.CalendarSubscription{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, family_id, name, url, created_by, created_at, last_synced_at, last_sync_error
		 FROM calendar_subscriptions WHERE id = $1`,
		id,
	).Scan(&s.ID, &s.FamilyID, &s.Name, &s.URL, &s.CreatedBy, &s.CreatedAt, &s.LastSyncedAt, &s.LastSyncError)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (r *CalendarSubscriptionRepository) Delete(ctx context.Context, id, familyID string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM calendar_subscriptions WHERE id = $1 AND family_id = $2`, id, familyID)
	return err
}

func (r *CalendarSubscriptionRepository) ListAll(ctx context.Context) ([]*model.CalendarSubscription, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, family_id, name, url, created_by, created_at, last_synced_at, last_sync_error
		 FROM calendar_subscriptions ORDER BY created_at`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*model.CalendarSubscription
	for rows.Next() {
		s := &model.CalendarSubscription{}
		if err := rows.Scan(&s.ID, &s.FamilyID, &s.Name, &s.URL, &s.CreatedBy, &s.CreatedAt, &s.LastSyncedAt, &s.LastSyncError); err != nil {
			return nil, err
		}
		result = append(result, s)
	}
	return result, rows.Err()
}

func (r *CalendarSubscriptionRepository) UpdateSyncResult(ctx context.Context, id string, syncErr *string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE calendar_subscriptions SET last_synced_at = NOW(), last_sync_error = $2 WHERE id = $1`,
		id, syncErr,
	)
	return err
}
