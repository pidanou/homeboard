package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/pidanou/homeboard/internal/service"
)

type CalendarExportHandler struct {
	export   *service.CalendarExportService
	families *service.HouseholdService
}

func NewCalendarExportHandler(export *service.CalendarExportService, families *service.HouseholdService) *CalendarExportHandler {
	return &CalendarExportHandler{export: export, families: families}
}

// Routes is mounted at /households/{familyID}/calendar/export-token, authenticated + admin-gated.
func (h *CalendarExportHandler) Routes() http.Handler {
	r := chi.NewRouter()
	r.Get("/", h.get)
	r.Post("/", h.regenerate)
	r.Delete("/", h.revoke)
	return r
}

// PublicRoutes is mounted at /calendar, unauthenticated.
func (h *CalendarExportHandler) PublicRoutes() http.Handler {
	r := chi.NewRouter()
	r.Get("/export/{token}.ics", h.serveICS)
	return r
}

func (h *CalendarExportHandler) get(w http.ResponseWriter, r *http.Request) {
	familyID := chi.URLParam(r, "familyID")
	if err := requireAdmin(r, familyID, h.families); err != nil {
		writeError(w, http.StatusForbidden, codeFor(err, "forbidden"))
		return
	}
	token := h.export.GetOrNil(r.Context(), familyID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(token)
}

func (h *CalendarExportHandler) regenerate(w http.ResponseWriter, r *http.Request) {
	familyID := chi.URLParam(r, "familyID")
	if err := requireAdmin(r, familyID, h.families); err != nil {
		writeError(w, http.StatusForbidden, codeFor(err, "forbidden"))
		return
	}
	userID := r.Context().Value(ContextKeyUserID).(string)

	token, err := h.export.Regenerate(r.Context(), familyID, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "generate_export_token_failed")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(token)
}

func (h *CalendarExportHandler) revoke(w http.ResponseWriter, r *http.Request) {
	familyID := chi.URLParam(r, "familyID")
	if err := requireAdmin(r, familyID, h.families); err != nil {
		writeError(w, http.StatusForbidden, codeFor(err, "forbidden"))
		return
	}
	if err := h.export.Revoke(r.Context(), familyID); err != nil {
		writeError(w, http.StatusInternalServerError, "revoke_export_token_failed")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *CalendarExportHandler) serveICS(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	data, err := h.export.BuildICS(r.Context(), token)
	if err != nil {
		writeError(w, http.StatusNotFound, "export_not_found")
		return
	}
	w.Header().Set("Content-Type", "text/calendar; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="calendar.ics"`)
	w.Write(data)
}
