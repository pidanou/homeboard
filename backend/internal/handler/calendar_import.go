package handler

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/pidanou/homeboard/internal/service"
)

const maxImportBytes = 5 << 20 // 5MB

type CalendarImportHandler struct {
	importSvc *service.CalendarImportService
	families  *service.HouseholdService
}

func NewCalendarImportHandler(importSvc *service.CalendarImportService, families *service.HouseholdService) *CalendarImportHandler {
	return &CalendarImportHandler{importSvc: importSvc, families: families}
}

// Routes is mounted at /households/{familyID}/calendar/import, authenticated (member).
func (h *CalendarImportHandler) Routes() http.Handler {
	r := chi.NewRouter()
	r.Post("/", h.upload)
	return r
}

func (h *CalendarImportHandler) upload(w http.ResponseWriter, r *http.Request) {
	familyID := chi.URLParam(r, "familyID")
	if err := requireMember(r, familyID, h.families); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	userID := r.Context().Value(ContextKeyUserID).(string)

	r.Body = http.MaxBytesReader(w, r.Body, maxImportBytes)
	if err := r.ParseMultipartForm(maxImportBytes); err != nil {
		http.Error(w, "file too large or invalid upload", http.StatusBadRequest)
		return
	}
	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "missing file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "failed to read file", http.StatusBadRequest)
		return
	}

	imported, skipped, err := h.importSvc.ImportFile(r.Context(), familyID, userID, data)
	if err != nil {
		http.Error(w, "failed to import calendar", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{"imported": imported, "skipped": skipped})
}
