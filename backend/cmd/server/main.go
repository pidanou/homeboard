package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pidanou/family-board/internal/handler"
	"github.com/pidanou/family-board/internal/repository/postgres"
	"github.com/pidanou/family-board/internal/service"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		slog.Error("DATABASE_URL is required")
		os.Exit(1)
	}

	m, err := migrate.New("file://migrations", dbURL)
	if err != nil {
		slog.Error("failed to init migrations", "err", err)
		os.Exit(1)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		slog.Error("failed to run migrations", "err", err)
		os.Exit(1)
	}

	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		slog.Error("failed to connect to database", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	// Repositories
	userRepo := postgres.NewUserRepository(pool)
	familyRepo := postgres.NewFamilyRepository(pool)
	inviteRepo := postgres.NewInviteRepository(pool)
	taskRepo := postgres.NewTaskRepository(pool)
	eventRepo := postgres.NewEventRepository(pool)
	labelRepo := postgres.NewCategoryRepository(pool)
	listRepo := postgres.NewListRepository(pool)

	// Services
	authService := service.NewAuthService(userRepo, os.Getenv("JWT_SECRET"))
	familyService := service.NewFamilyService(familyRepo)
	inviteService := service.NewInviteService(inviteRepo, familyRepo)
	taskService := service.NewTaskService(taskRepo)
	eventService := service.NewEventService(eventRepo)
	labelService := service.NewCategoryService(labelRepo)
	listService := service.NewListService(listRepo)

	// SSE hub
	hub := handler.NewHub()

	uploadDir := os.Getenv("UPLOAD_DIR")
	if uploadDir == "" {
		uploadDir = "./uploads"
	}
	appBaseURL := os.Getenv("APP_BASE_URL")

	// Handlers
	authHandler := handler.NewAuthHandler(authService)
	profileHandler := handler.NewProfileHandler(authService, uploadDir, appBaseURL)
	familyHandler := handler.NewFamilyHandler(familyService)
	inviteHandler := handler.NewInviteHandler(inviteService, os.Getenv("JWT_SECRET"))
	taskHandler := handler.NewTaskHandler(taskService, hub)
	eventHandler := handler.NewEventHandler(eventService, hub)
	labelHandler := handler.NewCategoryHandler(labelService, hub)
	listHandler := handler.NewListHandler(listService, hub)
	sseHandler := handler.NewSSEHandler(hub, os.Getenv("JWT_SECRET"))

	allowedOrigins := []string{"http://localhost:5173", "https://*.ngrok-free.app", "https://*.ngrok.io"}
	if origin := os.Getenv("APP_BASE_URL"); origin != "" {
		allowedOrigins = append(allowedOrigins, origin)
	}

	r := chi.NewRouter()
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowCredentials: false,
		MaxAge:           300,
	}))
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/api/v1", func(r chi.Router) {
		r.Mount("/auth", authHandler.Routes())
		r.Mount("/invites", inviteHandler.PublicRoutes())
		// SSE stream: does its own JWT auth via ?token= (EventSource can't set headers)
		r.Route("/families/{familyID}/stream", func(r chi.Router) {
			r.Mount("/", sseHandler.Routes())
		})
		r.Handle("/uploads/avatars/*", http.StripPrefix("/api/v1/uploads/avatars/", http.FileServer(http.Dir(uploadDir+"/avatars"))))

		r.Group(func(r chi.Router) {
			r.Use(handler.AuthMiddleware(os.Getenv("JWT_SECRET")))
			r.Mount("/profile", profileHandler.Routes())
			r.Mount("/families", familyHandler.Routes())
			r.Route("/families/{familyID}/invites", func(r chi.Router) {
				r.Mount("/", inviteHandler.Routes())
			})
			r.Route("/families/{familyID}/tasks", func(r chi.Router) {
				r.Mount("/", taskHandler.Routes())
			})
			r.Route("/families/{familyID}/events", func(r chi.Router) {
				r.Mount("/", eventHandler.Routes())
			})
			r.Route("/families/{familyID}/categories", func(r chi.Router) {
				r.Mount("/", labelHandler.Routes())
			})
			r.Route("/families/{familyID}/lists", func(r chi.Router) {
				r.Mount("/", listHandler.Routes())
			})
		})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	slog.Info("server starting", "port", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), r); err != nil {
		slog.Error("server error", "err", err)
		os.Exit(1)
	}
}
