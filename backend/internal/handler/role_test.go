package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pidanou/homeboard/internal/handler"
	"github.com/pidanou/homeboard/internal/model"
	pgRepo "github.com/pidanou/homeboard/internal/repository/postgres"
	"github.com/pidanou/homeboard/internal/service"
)

const testJWTSecret = "test-secret-for-role-enforcement"

// ── helpers ──────────────────────────────────────────────────────────────────

type testEnv struct {
	server *httptest.Server
	pool   *pgxpool.Pool
}

func newTestEnv(t *testing.T) *testEnv {
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

	householdRepo := pgRepo.NewHouseholdRepository(pool)
	inviteRepo := pgRepo.NewInviteRepository(pool)
	labelRepo := pgRepo.NewCategoryRepository(pool)

	householdSvc := service.NewHouseholdService(householdRepo)
	inviteSvc := service.NewInviteService(inviteRepo, householdRepo)
	labelSvc := service.NewCategoryService(labelRepo)
	hub := handler.NewHub()

	householdH := handler.NewHouseholdHandler(householdSvc)
	inviteH := handler.NewInviteHandler(inviteSvc, householdSvc, testJWTSecret)
	labelH := handler.NewCategoryHandler(labelSvc, householdSvc, hub)

	r := chi.NewRouter()
	r.Group(func(r chi.Router) {
		r.Use(handler.AuthMiddleware(testJWTSecret))
		r.Mount("/families", householdH.Routes())
		r.Route("/households/{familyID}/invites", func(r chi.Router) {
			r.Mount("/", inviteH.Routes())
		})
		r.Route("/households/{familyID}/categories", func(r chi.Router) {
			r.Mount("/", labelH.Routes())
		})
	})

	srv := httptest.NewServer(r)
	t.Cleanup(func() {
		srv.Close()
		pool.Close()
	})
	return &testEnv{server: srv, pool: pool}
}

// seedFamily creates a family and two members directly in the DB.
// Returns (familyID, adminUserID, memberUserID).
func (e *testEnv) seedFamily(t *testing.T) (familyID, adminID, memberID string) {
	t.Helper()
	ctx := context.Background()
	familyID = uuid.NewString()
	adminID = uuid.NewString()
	memberID = uuid.NewString()

	// Insert dummy users (just the id — foreign key requires users table)
	for _, u := range []struct{ id, email, name string }{
		{adminID, fmt.Sprintf("admin-%s@test.com", adminID[:8]), "Test Admin"},
		{memberID, fmt.Sprintf("member-%s@test.com", memberID[:8]), "Test Member"},
	} {
		_, err := e.pool.Exec(ctx,
			`INSERT INTO users (id, email, name, password_hash, created_at) VALUES ($1, $2, $3, 'x', $4)
			 ON CONFLICT (id) DO NOTHING`,
			u.id, u.email, u.name, time.Now().UTC(),
		)
		if err != nil {
			t.Fatalf("insert user: %v", err)
		}
	}

	_, err := e.pool.Exec(ctx,
		`INSERT INTO households (id, name, created_at) VALUES ($1, $2, $3)`,
		familyID, "Test Family "+familyID[:8], time.Now().UTC(),
	)
	if err != nil {
		t.Fatalf("insert family: %v", err)
	}

	for _, m := range []struct{ id, role string }{{adminID, "admin"}, {memberID, "member"}} {
		_, err := e.pool.Exec(ctx,
			`INSERT INTO household_members (family_id, user_id, role, joined_at) VALUES ($1, $2, $3, $4)`,
			familyID, m.id, m.role, time.Now().UTC(),
		)
		if err != nil {
			t.Fatalf("insert member: %v", err)
		}
	}

	t.Cleanup(func() {
		e.pool.Exec(ctx, `DELETE FROM household_members WHERE family_id = $1`, familyID)
		e.pool.Exec(ctx, `DELETE FROM households WHERE id = $1`, familyID)
		e.pool.Exec(ctx, `DELETE FROM users WHERE id = ANY($1)`, []string{adminID, memberID})
	})

	return familyID, adminID, memberID
}

func (e *testEnv) token(userID string) string {
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(time.Hour).Unix(),
	})
	tok, _ := t.SignedString([]byte(testJWTSecret))
	return tok
}

func (e *testEnv) do(method, url string, token string, body any) *http.Response {
	var buf bytes.Buffer
	if body != nil {
		json.NewEncoder(&buf).Encode(body)
	}
	req, _ := http.NewRequest(method, e.server.URL+url, &buf)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, _ := http.DefaultClient.Do(req)
	return resp
}

// ── role enforcement tests ────────────────────────────────────────────────────

func TestKickMember(t *testing.T) {
	e := newTestEnv(t)
	familyID, adminID, memberID := e.seedFamily(t)
	url := fmt.Sprintf("/households/%s/members/%s", familyID, memberID)

	t.Run("member cannot kick", func(t *testing.T) {
		resp := e.do("DELETE", url, e.token(memberID), nil)
		if resp.StatusCode != http.StatusForbidden {
			t.Errorf("want 403, got %d", resp.StatusCode)
		}
	})

	t.Run("admin can kick", func(t *testing.T) {
		// Add a third user to kick so we don't break subsequent tests
		ctx := context.Background()
		targetID := uuid.NewString()
		e.pool.Exec(ctx, `INSERT INTO users (id, email, name, password_hash, created_at) VALUES ($1, $2, $3, 'x', $4)`,
			targetID, "kick-target@test.com", "Kick Target", time.Now().UTC())
		e.pool.Exec(ctx, `INSERT INTO household_members (family_id, user_id, role, joined_at) VALUES ($1, $2, 'member', $3)`,
			familyID, targetID, time.Now().UTC())
		t.Cleanup(func() {
			e.pool.Exec(ctx, `DELETE FROM household_members WHERE user_id = $1`, targetID)
			e.pool.Exec(ctx, `DELETE FROM users WHERE id = $1`, targetID)
		})

		resp := e.do("DELETE", fmt.Sprintf("/households/%s/members/%s", familyID, targetID), e.token(adminID), nil)
		if resp.StatusCode != http.StatusNoContent {
			t.Errorf("want 204, got %d", resp.StatusCode)
		}
	})
}

func TestUpdateRole(t *testing.T) {
	e := newTestEnv(t)
	familyID, adminID, memberID := e.seedFamily(t)
	url := fmt.Sprintf("/households/%s/members/%s/role", familyID, memberID)

	t.Run("member cannot change role", func(t *testing.T) {
		resp := e.do("PUT", url, e.token(memberID), model.HouseholdMember{Role: "admin"})
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("want 400, got %d", resp.StatusCode)
		}
	})

	t.Run("admin can promote member", func(t *testing.T) {
		resp := e.do("PUT", url, e.token(adminID), map[string]string{"role": "admin"})
		if resp.StatusCode != http.StatusNoContent {
			t.Errorf("want 204, got %d", resp.StatusCode)
		}
	})

	t.Run("cannot demote last admin", func(t *testing.T) {
		// adminID is now the only admin (memberID was promoted in previous subtest)
		// promote memberID back to check, then demote adminID should fail
		e.pool.Exec(context.Background(),
			`UPDATE household_members SET role = 'member' WHERE user_id = $1 AND family_id = $2`,
			memberID, familyID)
		resp := e.do("PUT", fmt.Sprintf("/households/%s/members/%s/role", familyID, adminID),
			e.token(adminID), map[string]string{"role": "member"})
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("want 400 (last admin protection), got %d", resp.StatusCode)
		}
	})
}

func TestCreateVirtualMember(t *testing.T) {
	e := newTestEnv(t)
	familyID, adminID, memberID := e.seedFamily(t)
	url := fmt.Sprintf("/households/%s/members/virtual", familyID)
	body := map[string]string{"name": "Virtual Kid"}

	t.Run("member cannot create virtual", func(t *testing.T) {
		resp := e.do("POST", url, e.token(memberID), body)
		if resp.StatusCode != http.StatusForbidden {
			t.Errorf("want 403, got %d", resp.StatusCode)
		}
	})

	t.Run("admin can create virtual", func(t *testing.T) {
		resp := e.do("POST", url, e.token(adminID), body)
		if resp.StatusCode != http.StatusCreated {
			t.Errorf("want 201, got %d", resp.StatusCode)
		}
		var vm model.VirtualMember
		json.NewDecoder(resp.Body).Decode(&vm)
		t.Cleanup(func() {
			e.pool.Exec(context.Background(), `DELETE FROM virtual_members WHERE id = $1`, vm.ID)
		})
	})
}

func TestCategoryMutations(t *testing.T) {
	e := newTestEnv(t)
	familyID, adminID, memberID := e.seedFamily(t)

	// Seed a category for update/delete tests
	catID := uuid.NewString()
	e.pool.Exec(context.Background(),
		`INSERT INTO categories (id, family_id, name, color, created_at) VALUES ($1, $2, 'Test', 'blue', $3)`,
		catID, familyID, time.Now().UTC())
	t.Cleanup(func() {
		e.pool.Exec(context.Background(), `DELETE FROM categories WHERE id = $1`, catID)
	})

	cases := []struct {
		name   string
		method string
		url    string
		body   any
	}{
		{"create", "POST", fmt.Sprintf("/households/%s/categories", familyID), map[string]string{"name": "X", "color": "red"}},
		{"update", "PUT", fmt.Sprintf("/households/%s/categories/%s", familyID, catID), map[string]string{"name": "Y", "color": "green"}},
		{"delete", "DELETE", fmt.Sprintf("/households/%s/categories/%s", familyID, catID), nil},
	}

	for _, tc := range cases {
		t.Run(tc.name+" member forbidden", func(t *testing.T) {
			resp := e.do(tc.method, tc.url, e.token(memberID), tc.body)
			if resp.StatusCode != http.StatusForbidden {
				t.Errorf("%s %s: want 403, got %d", tc.method, tc.url, resp.StatusCode)
			}
		})
	}

	t.Run("admin can create category", func(t *testing.T) {
		resp := e.do("POST", fmt.Sprintf("/households/%s/categories", familyID), e.token(adminID),
			map[string]string{"name": "AdminCat", "color": "teal"})
		if resp.StatusCode != http.StatusCreated {
			t.Errorf("want 201, got %d", resp.StatusCode)
		}
		var cat model.Category
		json.NewDecoder(resp.Body).Decode(&cat)
		t.Cleanup(func() {
			e.pool.Exec(context.Background(), `DELETE FROM categories WHERE id = $1`, cat.ID)
		})
	})
}

func TestInviteMutations(t *testing.T) {
	e := newTestEnv(t)
	familyID, adminID, memberID := e.seedFamily(t)

	t.Run("member cannot generate invite", func(t *testing.T) {
		resp := e.do("POST", fmt.Sprintf("/households/%s/invites", familyID), e.token(memberID), map[string]string{})
		if resp.StatusCode != http.StatusForbidden {
			t.Errorf("want 403, got %d", resp.StatusCode)
		}
	})

	t.Run("admin can generate and revoke invite", func(t *testing.T) {
		resp := e.do("POST", fmt.Sprintf("/households/%s/invites", familyID), e.token(adminID), map[string]string{})
		if resp.StatusCode != http.StatusCreated {
			t.Errorf("want 201, got %d", resp.StatusCode)
		}
		var inv model.Invite
		json.NewDecoder(resp.Body).Decode(&inv)

		resp2 := e.do("DELETE", fmt.Sprintf("/households/%s/invites/%s", familyID, inv.Token), e.token(adminID), nil)
		if resp2.StatusCode != http.StatusNoContent {
			t.Errorf("revoke: want 204, got %d", resp2.StatusCode)
		}
	})

	t.Run("member cannot revoke invite", func(t *testing.T) {
		// admin generates one first
		resp := e.do("POST", fmt.Sprintf("/households/%s/invites", familyID), e.token(adminID), map[string]string{})
		var inv model.Invite
		json.NewDecoder(resp.Body).Decode(&inv)
		t.Cleanup(func() {
			e.pool.Exec(context.Background(), `DELETE FROM invites WHERE token = $1`, inv.Token)
		})

		resp2 := e.do("DELETE", fmt.Sprintf("/households/%s/invites/%s", familyID, inv.Token), e.token(memberID), nil)
		if resp2.StatusCode != http.StatusForbidden {
			t.Errorf("want 403, got %d", resp2.StatusCode)
		}
	})
}
