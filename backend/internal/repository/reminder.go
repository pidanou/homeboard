package repository

import (
	"context"
	"time"
)

type ReminderRecipient struct {
	UserID        string
	FamilyID      string
	MinutesBefore int
}

type ReminderRepository interface {
	// ListRecipients returns one row per (user, family) pair for every user
	// with reminders enabled (users.reminder_minutes_before IS NOT NULL).
	ListRecipients(ctx context.Context) ([]*ReminderRecipient, error)
	HasSent(ctx context.Context, userID, itemType, itemID string, occurrenceAt time.Time) (bool, error)
	MarkSent(ctx context.Context, userID, itemType, itemID string, occurrenceAt time.Time) error
}
