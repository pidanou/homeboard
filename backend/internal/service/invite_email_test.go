package service

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pidanou/homeboard/internal/repository/postgres"
)

func newInviteEmailTestEnv(t *testing.T) (*InviteService, *pgxpool.Pool, string, string) {
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

	ctx := context.Background()
	familyID := uuid.NewString()
	userID := uuid.NewString()
	now := time.Now().UTC()

	if _, err := pool.Exec(ctx, `INSERT INTO users (id, email, name, created_at) VALUES ($1, $2, $3, $4)`,
		userID, fmt.Sprintf("invite-admin-%s@test.com", userID[:8]), "Admin", now); err != nil {
		t.Fatalf("insert user: %v", err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO households (id, name, created_at) VALUES ($1, $2, $3)`,
		familyID, "Test Household", now); err != nil {
		t.Fatalf("insert household: %v", err)
	}
	t.Cleanup(func() {
		pool.Exec(context.Background(), `DELETE FROM households WHERE id = $1`, familyID)
		pool.Exec(context.Background(), `DELETE FROM users WHERE id = $1`, userID)
	})

	svc := NewInviteService(postgres.NewInviteRepository(pool), postgres.NewHouseholdRepository(pool), nil, "http://localhost:5173")
	return svc, pool, familyID, userID
}

func TestCreateInviteStoresEmail(t *testing.T) {
	svc, pool, familyID, userID := newInviteEmailTestEnv(t)

	invite, err := svc.Create(context.Background(), familyID, userID, "invitee@test.com")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if invite.Email == nil || *invite.Email != "invitee@test.com" {
		t.Fatalf("expected email to be stored on invite, got %+v", invite.Email)
	}

	got, err := svc.GetByToken(context.Background(), invite.Token)
	if err != nil {
		t.Fatalf("GetByToken: %v", err)
	}
	if got.Email == nil || *got.Email != "invitee@test.com" {
		t.Errorf("expected email to round-trip through storage, got %+v", got.Email)
	}
	_ = pool
}

func TestCreateInviteWithoutEmailLeavesItNil(t *testing.T) {
	svc, _, familyID, userID := newInviteEmailTestEnv(t)

	invite, err := svc.Create(context.Background(), familyID, userID, "")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if invite.Email != nil {
		t.Errorf("expected no email on a plain link invite, got %v", *invite.Email)
	}
}

func TestResendInviteRequiresStoredEmail(t *testing.T) {
	svc, _, familyID, userID := newInviteEmailTestEnv(t)

	invite, err := svc.Create(context.Background(), familyID, userID, "")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := svc.Resend(context.Background(), invite.Token); err == nil {
		t.Error("expected resend to fail for an invite with no email on file")
	}
}

func TestResendInviteWithEmailSucceeds(t *testing.T) {
	svc, _, familyID, userID := newInviteEmailTestEnv(t)

	invite, err := svc.Create(context.Background(), familyID, userID, "invitee@test.com")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := svc.Resend(context.Background(), invite.Token); err != nil {
		t.Errorf("expected resend to succeed, got: %v", err)
	}
}
