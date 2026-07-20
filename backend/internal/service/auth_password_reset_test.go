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

func newPasswordResetTestEnv(t *testing.T) (*AuthService, *pgxpool.Pool) {
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

	svc := NewAuthService(postgres.NewUserRepository(pool), postgres.NewPasswordResetRepository(pool), "test-secret", nil, "http://localhost:5173")
	return svc, pool
}

func insertPasswordUser(t *testing.T, pool *pgxpool.Pool, email string) string {
	t.Helper()
	id := uuid.NewString()
	hash := "$2a$10$abcdefghijklmnopqrstuuJhV8Z8nQ0z8yq0y8y8y8y8y8y8y8y8y" // not a real bcrypt hash, only used to mark the account password-based
	if _, err := pool.Exec(context.Background(),
		`INSERT INTO users (id, email, name, password_hash, created_at) VALUES ($1, $2, $3, $4, $5)`,
		id, email, "Reset Test User", hash, time.Now().UTC(),
	); err != nil {
		t.Fatalf("insert user: %v", err)
	}
	t.Cleanup(func() {
		pool.Exec(context.Background(), `DELETE FROM users WHERE id = $1`, id)
	})
	return id
}

func TestRequestPasswordResetCreatesSingleUseToken(t *testing.T) {
	svc, pool := newPasswordResetTestEnv(t)
	email := fmt.Sprintf("reset-%s@test.com", uuid.NewString()[:8])
	userID := insertPasswordUser(t, pool, email)

	if err := svc.RequestPasswordReset(context.Background(), email); err != nil {
		t.Fatalf("RequestPasswordReset: %v", err)
	}

	var token string
	if err := pool.QueryRow(context.Background(),
		`SELECT token FROM password_reset_tokens WHERE user_id = $1`, userID,
	).Scan(&token); err != nil {
		t.Fatalf("expected a reset token row: %v", err)
	}

	if err := svc.ResetPassword(context.Background(), token, "a-new-password123"); err != nil {
		t.Fatalf("ResetPassword: %v", err)
	}

	// Token is single-use.
	if err := svc.ResetPassword(context.Background(), token, "another-password123"); err == nil {
		t.Error("expected reused token to be rejected")
	}
}

func TestRequestPasswordResetUnknownEmailIsSilentNoOp(t *testing.T) {
	svc, _ := newPasswordResetTestEnv(t)
	if err := svc.RequestPasswordReset(context.Background(), "does-not-exist@test.com"); err != nil {
		t.Fatalf("expected no error for unknown email (avoid enumeration), got: %v", err)
	}
}

func TestResetPasswordRejectsExpiredToken(t *testing.T) {
	svc, pool := newPasswordResetTestEnv(t)
	email := fmt.Sprintf("reset-expired-%s@test.com", uuid.NewString()[:8])
	userID := insertPasswordUser(t, pool, email)

	token := uuid.NewString()
	if _, err := pool.Exec(context.Background(),
		`INSERT INTO password_reset_tokens (token, user_id, created_at, expires_at) VALUES ($1, $2, $3, $4)`,
		token, userID, time.Now().UTC().Add(-2*time.Hour), time.Now().UTC().Add(-time.Hour),
	); err != nil {
		t.Fatalf("insert expired token: %v", err)
	}

	if err := svc.ResetPassword(context.Background(), token, "a-new-password123"); err == nil {
		t.Error("expected expired token to be rejected")
	}
}
