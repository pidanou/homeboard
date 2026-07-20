package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pidanou/homeboard/internal/handler"
	"github.com/pidanou/homeboard/internal/model"
	pgRepo "github.com/pidanou/homeboard/internal/repository/postgres"
	"github.com/pidanou/homeboard/internal/service"
)

func newListTestEnv(t *testing.T) *testEnv {
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

	listRepo := pgRepo.NewListRepository(pool)
	listSvc := service.NewListService(listRepo)
	listH := handler.NewListHandler(listSvc, handler.NewHub())

	r := chi.NewRouter()
	r.Group(func(r chi.Router) {
		r.Use(handler.AuthMiddleware(testJWTSecret))
		r.Route("/households/{familyID}/lists", func(r chi.Router) {
			r.Mount("/", listH.Routes())
		})
	})

	srv := httptest.NewServer(r)
	t.Cleanup(func() {
		srv.Close()
		pool.Close()
	})
	return &testEnv{server: srv, pool: pool}
}

func TestListItemCategory(t *testing.T) {
	e := newListTestEnv(t)
	familyID, adminID, _ := e.seedFamily(t)
	tok := e.token(adminID)

	resp := e.do("POST", "/households/"+familyID+"/lists", tok, map[string]string{"name": "Shopping"})
	var list model.List
	json.NewDecoder(resp.Body).Decode(&list)

	resp = e.do("POST", "/households/"+familyID+"/lists/"+list.ID+"/items", tok, map[string]string{"name": "Milk"})
	var item model.ListItem
	json.NewDecoder(resp.Body).Decode(&item)

	t.Run("sets category", func(t *testing.T) {
		category := "Dairy"
		resp := e.do("PATCH", "/households/"+familyID+"/lists/"+list.ID+"/items/"+item.ID, tok,
			map[string]any{"name": item.Name, "checked": false, "category": category})
		if resp.StatusCode != http.StatusNoContent {
			t.Fatalf("want 204, got %d", resp.StatusCode)
		}

		var stored *string
		e.pool.QueryRow(context.Background(), `SELECT category FROM list_items WHERE id = $1`, item.ID).Scan(&stored)
		if stored == nil || *stored != category {
			t.Errorf("want category %q, got %v", category, stored)
		}
	})

	t.Run("clears category", func(t *testing.T) {
		resp := e.do("PATCH", "/households/"+familyID+"/lists/"+list.ID+"/items/"+item.ID, tok,
			map[string]any{"name": item.Name, "checked": false, "category": nil})
		if resp.StatusCode != http.StatusNoContent {
			t.Fatalf("want 204, got %d", resp.StatusCode)
		}

		var stored *string
		e.pool.QueryRow(context.Background(), `SELECT category FROM list_items WHERE id = $1`, item.ID).Scan(&stored)
		if stored != nil {
			t.Errorf("want cleared category, got %v", *stored)
		}
	})
}
