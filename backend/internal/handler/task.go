package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/pidanou/family-board/internal/service"
)

func parseOptionalTime(s *string) (*time.Time, error) {
	if s == nil || *s == "" {
		return nil, nil
	}
	t, err := time.Parse(time.RFC3339, *s)
	if err != nil {
		return nil, err
	}
	utc := t.UTC()
	return &utc, nil
}

type TaskHandler struct {
	tasks *service.TaskService
	hub   *Hub
}

func NewTaskHandler(tasks *service.TaskService, hub *Hub) *TaskHandler {
	return &TaskHandler{tasks: tasks, hub: hub}
}

func (h *TaskHandler) Routes() http.Handler {
	r := chi.NewRouter()
	r.Get("/", h.list)
	r.Post("/", h.create)
	r.Put("/reorder", h.reorder)
	r.Patch("/{taskID}", h.update)
	r.Delete("/{taskID}", h.delete)
	return r
}

func (h *TaskHandler) list(w http.ResponseWriter, r *http.Request) {
	familyID := chi.URLParam(r, "familyID")

	tasks, err := h.tasks.ListForFamily(r.Context(), familyID)
	if err != nil {
		http.Error(w, "failed to list tasks", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

func (h *TaskHandler) create(w http.ResponseWriter, r *http.Request) {
	familyID := chi.URLParam(r, "familyID")
	userID := r.Context().Value(ContextKeyUserID).(string)

	var body struct {
		Title       string  `json:"title"`
		Description string  `json:"description"`
		Important   bool    `json:"important"`
		AssignedTo  *string `json:"assigned_to"`
		StartDate   *string `json:"start_date"`
		EndDate     *string `json:"end_date"`
		CategoryID  *string `json:"category_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Title == "" {
		http.Error(w, "title is required", http.StatusBadRequest)
		return
	}

	startDate, err := parseOptionalTime(body.StartDate)
	if err != nil {
		http.Error(w, "invalid start_date format", http.StatusBadRequest)
		return
	}
	endDate, err := parseOptionalTime(body.EndDate)
	if err != nil {
		http.Error(w, "invalid end_date format", http.StatusBadRequest)
		return
	}

	task, err := h.tasks.Create(r.Context(), familyID, userID, body.Title, body.Description, body.Important, body.AssignedTo, startDate, endDate, body.CategoryID)
	if err != nil {
		http.Error(w, "failed to create task", http.StatusInternalServerError)
		return
	}

	h.hub.Broadcast(familyID)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(task)
}

func (h *TaskHandler) update(w http.ResponseWriter, r *http.Request) {
	familyID := chi.URLParam(r, "familyID")
	taskID := chi.URLParam(r, "taskID")

	task, err := h.tasks.GetByID(r.Context(), taskID, familyID)
	if err != nil {
		http.Error(w, "task not found", http.StatusNotFound)
		return
	}

	var body struct {
		Title       *string `json:"title"`
		Description *string `json:"description"`
		Important   *bool   `json:"important"`
		Status      *string `json:"status"`
		AssignedTo  *string `json:"assigned_to"`
		StartDate   *string `json:"start_date"`
		EndDate     *string `json:"end_date"`
		CategoryID  *string `json:"category_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if body.Title != nil {
		task.Title = *body.Title
	}
	if body.Description != nil {
		task.Description = *body.Description
	}
	if body.Important != nil {
		task.Important = *body.Important
	}
	if body.Status != nil {
		task.Status = *body.Status
	}
	if body.AssignedTo != nil {
		task.AssignedTo = body.AssignedTo
	}
	if body.StartDate != nil {
		startDate, err := parseOptionalTime(body.StartDate)
		if err != nil {
			http.Error(w, "invalid start_date format", http.StatusBadRequest)
			return
		}
		task.StartDate = startDate
	}
	if body.EndDate != nil {
		endDate, err := parseOptionalTime(body.EndDate)
		if err != nil {
			http.Error(w, "invalid end_date format", http.StatusBadRequest)
			return
		}
		task.EndDate = endDate
	}
	if body.CategoryID != nil {
		task.CategoryID = body.CategoryID
	}

	if err := h.tasks.Update(r.Context(), task); err != nil {
		http.Error(w, "failed to update task", http.StatusInternalServerError)
		return
	}

	h.hub.Broadcast(familyID)
	w.WriteHeader(http.StatusNoContent)
}

func (h *TaskHandler) reorder(w http.ResponseWriter, r *http.Request) {
	familyID := chi.URLParam(r, "familyID")
	var body struct {
		IDs []string `json:"ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || len(body.IDs) == 0 {
		http.Error(w, "ids is required", http.StatusBadRequest)
		return
	}
	if err := h.tasks.Reorder(r.Context(), familyID, body.IDs); err != nil {
		http.Error(w, "failed to reorder tasks", http.StatusInternalServerError)
		return
	}
	h.hub.Broadcast(familyID)
	w.WriteHeader(http.StatusNoContent)
}

func (h *TaskHandler) delete(w http.ResponseWriter, r *http.Request) {
	familyID := chi.URLParam(r, "familyID")
	taskID := chi.URLParam(r, "taskID")

	if err := h.tasks.Delete(r.Context(), taskID, familyID); err != nil {
		http.Error(w, "failed to delete task", http.StatusInternalServerError)
		return
	}

	h.hub.Broadcast(familyID)
	w.WriteHeader(http.StatusNoContent)
}
