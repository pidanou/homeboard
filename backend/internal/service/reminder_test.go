package service

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pidanou/homeboard/internal/model"
	"github.com/pidanou/homeboard/internal/repository/postgres"
)

func newReminderTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		dbURL = os.Getenv("DATABASE_URL")
	}
	if dbURL == "" {
		t.Skip("DATABASE_URL not set — skipping integration tests")
	}
	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		t.Fatalf("connect db: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}

// seedReminderFamily creates a household with two members: userA has a 15-minute
// reminder offset configured, userB has none.
func seedReminderFamily(t *testing.T, pool *pgxpool.Pool) (familyID, userA, userB string) {
	t.Helper()
	ctx := context.Background()
	familyID = uuid.NewString()
	userA = uuid.NewString()
	userB = uuid.NewString()
	now := time.Now().UTC()

	fifteen := 15
	if _, err := pool.Exec(ctx,
		`INSERT INTO users (id, email, name, password_hash, created_at, reminder_minutes_before) VALUES ($1, $2, $3, 'x', $4, $5)`,
		userA, fmt.Sprintf("reminder-a-%s@test.com", userA[:8]), "Reminder Test User A", now, fifteen,
	); err != nil {
		t.Fatalf("insert user A: %v", err)
	}
	if _, err := pool.Exec(ctx,
		`INSERT INTO users (id, email, name, password_hash, created_at) VALUES ($1, $2, $3, 'x', $4)`,
		userB, fmt.Sprintf("reminder-b-%s@test.com", userB[:8]), "Reminder Test User B", now,
	); err != nil {
		t.Fatalf("insert user B: %v", err)
	}
	if _, err := pool.Exec(ctx,
		`INSERT INTO households (id, name, created_at) VALUES ($1, $2, $3)`,
		familyID, "Reminder Test Family", now,
	); err != nil {
		t.Fatalf("insert household: %v", err)
	}
	if _, err := pool.Exec(ctx,
		`INSERT INTO household_members (family_id, user_id) VALUES ($1, $2), ($1, $3)`,
		familyID, userA, userB,
	); err != nil {
		t.Fatalf("insert household members: %v", err)
	}

	t.Cleanup(func() {
		pool.Exec(ctx, `DELETE FROM households WHERE id = $1`, familyID)
		pool.Exec(ctx, `DELETE FROM users WHERE id = $1`, userA)
		pool.Exec(ctx, `DELETE FROM users WHERE id = $1`, userB)
	})
	return familyID, userA, userB
}

func newReminderTestService(pool *pgxpool.Pool) (*ReminderService, *postgres.TaskRepository, *postgres.EventRepository) {
	taskRepo := postgres.NewTaskRepository(pool)
	eventRepo := postgres.NewEventRepository(pool)
	reminderRepo := postgres.NewReminderRepository(pool)
	pushService := NewPushService(postgres.NewPushRepository(pool), "", "", "")
	return NewReminderService(reminderRepo, taskRepo, eventRepo, pushService), taskRepo, eventRepo
}

func countSentReminders(t *testing.T, pool *pgxpool.Pool, userID, itemType, itemID string) int {
	t.Helper()
	var count int
	if err := pool.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM sent_reminders WHERE user_id = $1 AND item_type = $2 AND item_id = $3`,
		userID, itemType, itemID,
	).Scan(&count); err != nil {
		t.Fatalf("count sent_reminders: %v", err)
	}
	return count
}

func TestReminderService_TaskAssignedFiresOnceWithinWindow(t *testing.T) {
	pool := newReminderTestPool(t)
	familyID, userA, _ := seedReminderFamily(t, pool)
	reminderSvc, taskRepo, _ := newReminderTestService(pool)

	// +30s of slack past the 15-minute offset so the window match survives the
	// few ms of clock skew between "now" here and "now" inside CheckAndSend.
	dueAt := time.Now().UTC().Add(15*time.Minute + 30*time.Second)
	task := &model.Task{
		ID: uuid.NewString(), FamilyID: familyID, Title: "Assigned Task",
		Status: "todo", AssignedTo: &userA, EndDate: &dueAt,
		CreatedBy: userA, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
	}
	if err := taskRepo.Create(context.Background(), task); err != nil {
		t.Fatalf("create task: %v", err)
	}

	reminderSvc.CheckAndSend(context.Background())
	if got := countSentReminders(t, pool, userA, "task", task.ID); got != 1 {
		t.Fatalf("want 1 sent reminder after first check, got %d", got)
	}

	// Second tick must not duplicate the reminder.
	reminderSvc.CheckAndSend(context.Background())
	if got := countSentReminders(t, pool, userA, "task", task.ID); got != 1 {
		t.Fatalf("want 1 sent reminder after second check (idempotent), got %d", got)
	}
}

func TestReminderService_UnassignedTaskNeverFires(t *testing.T) {
	pool := newReminderTestPool(t)
	familyID, userA, _ := seedReminderFamily(t, pool)
	reminderSvc, taskRepo, _ := newReminderTestService(pool)

	dueAt := time.Now().UTC().Add(15*time.Minute + 30*time.Second)
	task := &model.Task{
		ID: uuid.NewString(), FamilyID: familyID, Title: "Unassigned Task",
		Status: "todo", EndDate: &dueAt,
		CreatedBy: userA, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
	}
	if err := taskRepo.Create(context.Background(), task); err != nil {
		t.Fatalf("create task: %v", err)
	}

	reminderSvc.CheckAndSend(context.Background())

	var count int
	if err := pool.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM sent_reminders WHERE item_type = 'task' AND item_id = $1`, task.ID,
	).Scan(&count); err != nil {
		t.Fatalf("count sent_reminders: %v", err)
	}
	if count != 0 {
		t.Fatalf("want 0 sent reminders for unassigned task, got %d", count)
	}
}

func TestReminderService_EventRespectsAttendeeFilter(t *testing.T) {
	pool := newReminderTestPool(t)
	familyID, userA, userB := seedReminderFamily(t, pool)
	// Give userB a reminder offset too, so we can prove the attendee filter (not
	// just the enabled/disabled flag) is what excludes them.
	if _, err := pool.Exec(context.Background(), `UPDATE users SET reminder_minutes_before = 15 WHERE id = $1`, userB); err != nil {
		t.Fatalf("enable reminders for userB: %v", err)
	}

	reminderSvc, _, eventRepo := newReminderTestService(pool)

	startAt := time.Now().UTC().Add(15*time.Minute + 30*time.Second)
	event := &model.Event{
		ID: uuid.NewString(), FamilyID: familyID, Title: "Attendee-only Event",
		StartAt: startAt, EndAt: startAt.Add(time.Hour), Type: "default",
		AttendeeIDs: []string{userA},
		CreatedBy:   userA, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
	}
	if err := eventRepo.Create(context.Background(), event); err != nil {
		t.Fatalf("create event: %v", err)
	}

	reminderSvc.CheckAndSend(context.Background())

	if got := countSentReminders(t, pool, userA, "event", event.ID); got != 1 {
		t.Fatalf("want 1 sent reminder for attendee userA, got %d", got)
	}
	if got := countSentReminders(t, pool, userB, "event", event.ID); got != 0 {
		t.Fatalf("want 0 sent reminders for non-attendee userB, got %d", got)
	}
}
