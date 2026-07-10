package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pidanou/homeboard/internal/model"
	"github.com/pidanou/homeboard/internal/repository/postgres"
)

const testIssuer = "https://idp.test.example.com"

func newOIDCTestEnv(t *testing.T, registrationGate func(context.Context) error) (*OIDCService, *pgxpool.Pool) {
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

	if registrationGate == nil {
		registrationGate = func(context.Context) error { return nil }
	}

	svc := &OIDCService{
		providerName:     "Test IdP",
		users:            postgres.NewUserRepository(pool),
		identities:       postgres.NewOIDCIdentityRepository(pool),
		issueToken:       func(userID string) (string, error) { return "token-for-" + userID, nil },
		registrationGate: registrationGate,
	}

	t.Cleanup(func() { pool.Close() })
	return svc, pool
}

func insertTestUser(t *testing.T, pool *pgxpool.Pool, email string) *model.User {
	t.Helper()
	user := &model.User{
		ID:        uuid.NewString(),
		Email:     email,
		Name:      "Existing User",
		Locale:    "en",
		CreatedAt: time.Now().UTC(),
	}
	_, err := pool.Exec(context.Background(),
		`INSERT INTO users (id, email, name, created_at) VALUES ($1, $2, $3, $4)`,
		user.ID, user.Email, user.Name, user.CreatedAt)
	if err != nil {
		t.Fatalf("insert test user: %v", err)
	}
	t.Cleanup(func() {
		pool.Exec(context.Background(), `DELETE FROM users WHERE id = $1`, user.ID)
	})
	return user
}

func TestFindOrCreateUserReturningIdentity(t *testing.T) {
	svc, pool := newOIDCTestEnv(t, nil)
	user := insertTestUser(t, pool, fmt.Sprintf("returning-%s@test.com", uuid.NewString()[:8]))

	_, err := pool.Exec(context.Background(),
		`INSERT INTO oidc_identities (id, user_id, issuer, subject, email, email_verified, created_at)
		 VALUES ($1, $2, $3, $4, $5, true, $6)`,
		uuid.NewString(), user.ID, testIssuer, "subject-1", user.Email, time.Now().UTC())
	if err != nil {
		t.Fatalf("insert identity: %v", err)
	}

	got, err := svc.FindOrCreateUser(context.Background(), OIDCClaims{
		Issuer: testIssuer, Subject: "subject-1", Email: user.Email, EmailVerified: false, // unverified now, shouldn't matter for a known identity
	})
	if err != nil {
		t.Fatalf("FindOrCreateUser: %v", err)
	}
	if got.ID != user.ID {
		t.Errorf("want existing user %s, got %s", user.ID, got.ID)
	}
}

func TestFindOrCreateUserAutoLinksVerifiedEmail(t *testing.T) {
	svc, pool := newOIDCTestEnv(t, nil)
	user := insertTestUser(t, pool, fmt.Sprintf("autolink-%s@test.com", uuid.NewString()[:8]))

	got, err := svc.FindOrCreateUser(context.Background(), OIDCClaims{
		Issuer: testIssuer, Subject: "subject-2", Email: user.Email, EmailVerified: true, Name: "Whoever",
	})
	if err != nil {
		t.Fatalf("FindOrCreateUser: %v", err)
	}
	if got.ID != user.ID {
		t.Errorf("want auto-linked to existing user %s, got %s", user.ID, got.ID)
	}

	identity, err := svc.identities.GetByIssuerSubject(context.Background(), testIssuer, "subject-2")
	if err != nil {
		t.Fatalf("expected identity to be linked: %v", err)
	}
	if identity.UserID != user.ID {
		t.Errorf("linked identity points at wrong user: %s", identity.UserID)
	}
}

func TestFindOrCreateUserRejectsUnverifiedEmail(t *testing.T) {
	svc, pool := newOIDCTestEnv(t, nil)
	user := insertTestUser(t, pool, fmt.Sprintf("unverified-%s@test.com", uuid.NewString()[:8]))

	_, err := svc.FindOrCreateUser(context.Background(), OIDCClaims{
		Issuer: testIssuer, Subject: "subject-3", Email: user.Email, EmailVerified: false,
	})
	if !errors.Is(err, ErrOIDCEmailNotVerified) {
		t.Errorf("want ErrOIDCEmailNotVerified, got %v", err)
	}

	if _, err := svc.identities.GetByIssuerSubject(context.Background(), testIssuer, "subject-3"); err == nil {
		t.Error("expected no identity to be created for an unverified email")
	}
}

func TestFindOrCreateUserCreatesNewUser(t *testing.T) {
	svc, pool := newOIDCTestEnv(t, nil)
	email := fmt.Sprintf("newuser-%s@test.com", uuid.NewString()[:8])
	subject := "subject-" + uuid.NewString()

	got, err := svc.FindOrCreateUser(context.Background(), OIDCClaims{
		Issuer: testIssuer, Subject: subject, Email: email, EmailVerified: true, Name: "New Person",
	})
	if err != nil {
		t.Fatalf("FindOrCreateUser: %v", err)
	}
	t.Cleanup(func() {
		pool.Exec(context.Background(), `DELETE FROM users WHERE id = $1`, got.ID)
	})
	if got.Email != email {
		t.Errorf("want email %s, got %s", email, got.Email)
	}
	if got.PasswordHash != nil {
		t.Error("want nil password hash for an OIDC-only account")
	}

	if _, err := svc.users.GetByEmail(context.Background(), email); err != nil {
		t.Fatalf("expected user to be persisted: %v", err)
	}
}

func TestFindOrCreateUserRespectsRegistrationGate(t *testing.T) {
	gateErr := ErrRegistrationClosed
	svc, _ := newOIDCTestEnv(t, func(context.Context) error { return gateErr })
	email := fmt.Sprintf("gated-%s@test.com", uuid.NewString()[:8])

	_, err := svc.FindOrCreateUser(context.Background(), OIDCClaims{
		Issuer: testIssuer, Subject: "subject-5", Email: email, EmailVerified: true,
	})
	if !errors.Is(err, ErrRegistrationClosed) {
		t.Errorf("want ErrRegistrationClosed, got %v", err)
	}
}
