package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pidanou/homeboard/internal/model"
)

type PushRepository struct {
	pool *pgxpool.Pool
}

func NewPushRepository(pool *pgxpool.Pool) *PushRepository {
	return &PushRepository{pool: pool}
}

func (r *PushRepository) Save(ctx context.Context, sub *model.PushSubscription) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO push_subscriptions (user_id, endpoint, auth, p256dh)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id, endpoint) DO UPDATE
		  SET auth = EXCLUDED.auth, p256dh = EXCLUDED.p256dh
	`, sub.UserID, sub.Endpoint, sub.Auth, sub.P256DH)
	return err
}

func (r *PushRepository) Delete(ctx context.Context, userID, endpoint string) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM push_subscriptions WHERE user_id = $1 AND endpoint = $2`,
		userID, endpoint)
	return err
}

func (r *PushRepository) ListForFamily(ctx context.Context, familyID string) ([]*model.PushSubscription, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT ps.id, ps.user_id, ps.endpoint, ps.auth, ps.p256dh
		FROM push_subscriptions ps
		JOIN household_members hm ON hm.user_id = ps.user_id
		WHERE hm.family_id = $1
	`, familyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subs []*model.PushSubscription
	for rows.Next() {
		s := &model.PushSubscription{}
		if err := rows.Scan(&s.ID, &s.UserID, &s.Endpoint, &s.Auth, &s.P256DH); err != nil {
			return nil, err
		}
		subs = append(subs, s)
	}
	return subs, rows.Err()
}

func (r *PushRepository) DeleteByEndpoint(ctx context.Context, endpoint string) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM push_subscriptions WHERE endpoint = $1`, endpoint)
	return err
}
