package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pidanou/homeboard/internal/repository"
)

type ReminderRepository struct {
	pool *pgxpool.Pool
}

func NewReminderRepository(pool *pgxpool.Pool) *ReminderRepository {
	return &ReminderRepository{pool: pool}
}

func (r *ReminderRepository) ListRecipients(ctx context.Context) ([]*repository.ReminderRecipient, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT hm.user_id, hm.family_id, u.reminder_minutes_before
		FROM household_members hm
		JOIN users u ON u.id = hm.user_id
		WHERE u.reminder_minutes_before IS NOT NULL
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var recipients []*repository.ReminderRecipient
	for rows.Next() {
		rec := &repository.ReminderRecipient{}
		if err := rows.Scan(&rec.UserID, &rec.FamilyID, &rec.MinutesBefore); err != nil {
			return nil, err
		}
		recipients = append(recipients, rec)
	}
	return recipients, rows.Err()
}

func (r *ReminderRepository) HasSent(ctx context.Context, userID, itemType, itemID string, occurrenceAt time.Time) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM sent_reminders
			WHERE user_id = $1 AND item_type = $2 AND item_id = $3 AND occurrence_at = $4
		)
	`, userID, itemType, itemID, occurrenceAt).Scan(&exists)
	return exists, err
}

func (r *ReminderRepository) MarkSent(ctx context.Context, userID, itemType, itemID string, occurrenceAt time.Time) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO sent_reminders (user_id, item_type, item_id, occurrence_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id, item_type, item_id, occurrence_at) DO NOTHING
	`, userID, itemType, itemID, occurrenceAt)
	return err
}
