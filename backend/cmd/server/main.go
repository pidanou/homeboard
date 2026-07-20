package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pidanou/homeboard/internal/config"
	"github.com/pidanou/homeboard/internal/handler"
	"github.com/pidanou/homeboard/internal/repository/postgres"
	"github.com/pidanou/homeboard/internal/service"
)

func main() {
	// pgx decodes timestamptz via time.Unix(), which Go always tags with the process's
	// time.Local — regardless of DB session timezone. On a non-UTC host that leaks
	// "+02:00"-style offsets into JSON responses instead of "Z", which breaks all-day
	// event date math on the frontend calendar. CLAUDE.md: all dates are UTC ISO 8601.
	time.Local = time.UTC

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
	oidcIdentityRepo := postgres.NewOIDCIdentityRepository(pool)
	householdRepo := postgres.NewHouseholdRepository(pool)
	inviteRepo := postgres.NewInviteRepository(pool)
	taskRepo := postgres.NewTaskRepository(pool)
	eventRepo := postgres.NewEventRepository(pool)
	labelRepo := postgres.NewCategoryRepository(pool)
	listRepo := postgres.NewListRepository(pool)
	pushRepo := postgres.NewPushRepository(pool)
	reminderRepo := postgres.NewReminderRepository(pool)
	calendarExportTokenRepo := postgres.NewCalendarExportTokenRepository(pool)
	calendarSubscriptionRepo := postgres.NewCalendarSubscriptionRepository(pool)

	// Services
	mailer := service.NewEmailService(
		os.Getenv("SMTP_HOST"),
		os.Getenv("SMTP_PORT"),
		os.Getenv("SMTP_USER"),
		os.Getenv("SMTP_PASS"),
		os.Getenv("SMTP_FROM"),
		os.Getenv("SMTP_TLS") == "true",
	)
	authConfig := config.LoadAuth()
	authService := service.NewAuthService(userRepo, authConfig.JWTSecret, mailer)
	authService.SetAllowPasswordLogin(authConfig.AllowPasswordLogin)

	var oidcHandler *handler.OIDCHandler
	oidcProviderName := ""
	if authConfig.OIDC != nil {
		oidcService, err := service.NewOIDCService(context.Background(), *authConfig.OIDC, userRepo, oidcIdentityRepo, authService.IssueToken, authService.CheckRegistrationOpen)
		if err != nil {
			slog.Warn("oidc discovery failed, starting with OIDC disabled", "issuer", authConfig.OIDC.IssuerURL, "err", err)
		} else {
			handoffStore := service.NewOIDCHandoffStore()
			oidcHandler = handler.NewOIDCHandler(oidcService, handoffStore, authConfig.JWTSecret, os.Getenv("APP_BASE_URL"), authConfig.OIDC.RedirectURL)
			oidcProviderName = authConfig.OIDC.ProviderName
		}
	}

	householdService := service.NewHouseholdService(householdRepo)
	inviteService := service.NewInviteService(inviteRepo, householdRepo)
	taskService := service.NewTaskService(taskRepo)
	eventService := service.NewEventService(eventRepo)
	labelService := service.NewCategoryService(labelRepo)
	listService := service.NewListService(listRepo)
	pushService := service.NewPushService(pushRepo, os.Getenv("VAPID_PRIVATE_KEY"), os.Getenv("VAPID_PUBLIC_KEY"), os.Getenv("VAPID_SUBJECT"))
	calendarExportService := service.NewCalendarExportService(calendarExportTokenRepo, eventRepo, householdRepo)
	calendarImportService := service.NewCalendarImportService(eventService)
	calendarSyncService := service.NewCalendarSyncService(calendarSubscriptionRepo, eventRepo)
	reminderService := service.NewReminderService(reminderRepo, taskRepo, eventRepo, pushService)

	// SSE hub
	hub := handler.NewHub()

	uploadDir := os.Getenv("UPLOAD_DIR")
	if uploadDir == "" {
		uploadDir = "./uploads"
	}
	// Handlers
	authHandler := handler.NewAuthHandler(authService)
	profileHandler := handler.NewProfileHandler(authService, uploadDir, os.Getenv("API_BASE_URL"))
	householdHandler := handler.NewHouseholdHandler(householdService, uploadDir, os.Getenv("API_BASE_URL"))
	inviteHandler := handler.NewInviteHandler(inviteService, householdService, authService, os.Getenv("JWT_SECRET"))
	taskHandler := handler.NewTaskHandler(taskService, hub, pushService)
	eventHandler := handler.NewEventHandler(eventService, hub, pushService)
	labelHandler := handler.NewCategoryHandler(labelService, householdService, hub)
	listHandler := handler.NewListHandler(listService, hub)
	sseHandler := handler.NewSSEHandler(hub, os.Getenv("JWT_SECRET"), householdService)
	pushHandler := handler.NewPushHandler(pushService, os.Getenv("VAPID_PUBLIC_KEY"))
	calendarExportHandler := handler.NewCalendarExportHandler(calendarExportService, householdService)
	calendarImportHandler := handler.NewCalendarImportHandler(calendarImportService, householdService)
	calendarSubscriptionHandler := handler.NewCalendarSubscriptionHandler(calendarSyncService, householdService)

	allowedOrigins := []string{"http://localhost:5173"}
	if extra := strings.TrimSpace(os.Getenv("CORS_ALLOWED_ORIGINS")); extra == "*" {
		allowedOrigins = []string{"*"}
	} else if extra != "" {
		for _, o := range strings.Split(extra, ",") {
			if o = strings.TrimSpace(o); o != "" {
				allowedOrigins = append(allowedOrigins, o)
			}
		}
	}
	if origin := os.Getenv("APP_BASE_URL"); origin != "" && allowedOrigins[0] != "*" {
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
	r.Use(handler.SecurityHeaders)

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/config", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"allowMultiHousehold": os.Getenv("ALLOW_MULTI_HOUSEHOLD") == "true",
				"supportedLocales":    []string{"en", "fr", "es", "de"},
				"oidcEnabled":         oidcHandler != nil,
				"oidcProviderName":    oidcProviderName,
				"allowPasswordLogin":  authConfig.AllowPasswordLogin,
			})
		})
		r.Mount("/auth", authHandler.Routes())
		if oidcHandler != nil {
			r.Mount("/auth/oidc", oidcHandler.Routes())
		}
		r.Mount("/invites", inviteHandler.PublicRoutes())
		r.Mount("/calendar", calendarExportHandler.PublicRoutes())
		// SSE stream: does its own JWT auth via ?token= (EventSource can't set headers)
		r.Route("/households/{familyID}/stream", func(r chi.Router) {
			r.Mount("/", sseHandler.Routes())
		})
		r.Group(func(r chi.Router) {
			r.Use(handler.AuthMiddleware(os.Getenv("JWT_SECRET")))
			r.Handle("/uploads/avatars/*", http.StripPrefix("/api/v1/uploads/avatars/", http.FileServer(http.Dir(uploadDir+"/avatars"))))
			r.Handle("/uploads/household/*", http.StripPrefix("/api/v1/uploads/household/", http.FileServer(http.Dir(uploadDir+"/household"))))
			r.Mount("/profile", profileHandler.Routes())
			r.Mount("/households", householdHandler.Routes())
			r.Route("/households/{familyID}/invites", func(r chi.Router) {
				r.Use(handler.RequireFamilyMember(householdService))
				r.Mount("/", inviteHandler.Routes())
			})
			r.Route("/households/{familyID}/tasks", func(r chi.Router) {
				r.Use(handler.RequireFamilyMember(householdService))
				r.Mount("/", taskHandler.Routes())
			})
			r.Route("/households/{familyID}/events", func(r chi.Router) {
				r.Use(handler.RequireFamilyMember(householdService))
				r.Mount("/", eventHandler.Routes())
			})
			r.Route("/households/{familyID}/categories", func(r chi.Router) {
				r.Use(handler.RequireFamilyMember(householdService))
				r.Mount("/", labelHandler.Routes())
			})
			r.Route("/households/{familyID}/lists", func(r chi.Router) {
				r.Use(handler.RequireFamilyMember(householdService))
				r.Mount("/", listHandler.Routes())
			})
			r.Route("/households/{familyID}/calendar/export-token", func(r chi.Router) {
				r.Use(handler.RequireFamilyMember(householdService))
				r.Mount("/", calendarExportHandler.Routes())
			})
			r.Route("/households/{familyID}/calendar/import", func(r chi.Router) {
				r.Use(handler.RequireFamilyMember(householdService))
				r.Mount("/", calendarImportHandler.Routes())
			})
			r.Route("/households/{familyID}/calendar/subscriptions", func(r chi.Router) {
				r.Use(handler.RequireFamilyMember(householdService))
				r.Mount("/", calendarSubscriptionHandler.Routes())
			})
			r.Mount("/push", pushHandler.Routes())
		})
	})

	syncInterval := 1 * time.Hour
	if v := os.Getenv("CALENDAR_SYNC_INTERVAL"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			syncInterval = d
		} else {
			slog.Warn("invalid CALENDAR_SYNC_INTERVAL, using default", "value", v, "err", err)
		}
	}
	if syncInterval < 5*time.Minute {
		syncInterval = 5 * time.Minute
	}
	go func() {
		ticker := time.NewTicker(syncInterval)
		defer ticker.Stop()
		for range ticker.C {
			calendarSyncService.SyncAll(context.Background())
		}
	}()

	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			reminderService.CheckAndSend(context.Background())
		}
	}()

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
