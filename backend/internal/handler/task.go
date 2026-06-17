package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/pidanou/family-board/internal/model"
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
		Title       string   `json:"title"`
		Description string   `json:"description"`
		Priority    string   `json:"priority"`
		AssignedTo  *string  `json:"assigned_to"`
		StartDate   *string  `json:"start_date"`
		EndDate     *string  `json:"end_date"`
		LabelIDs    []string `json:"label_ids"`
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

	task, err := h.tasks.Create(r.Context(), familyID, userID, body.Title, body.Description, body.Priority, body.AssignedTo, startDate, endDate, body.LabelIDs)
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
	taskID := chi.URLParam(r, "taskID")

	var body struct {
		Title       *string  `json:"title"`
		Description *string  `json:"description"`
		Priority    *string  `json:"priority"`
		Status      *string  `json:"status"`
		AssignedTo  *string  `json:"assigned_to"`
		StartDate   *string  `json:"start_date"`
		EndDate     *string  `json:"end_date"`
		LabelIDs    []string `json:"label_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	familyID := chi.URLParam(r, "familyID")
	task := &model.Task{ID: taskID, FamilyID: familyID}
	if body.Title != nil {
		task.Title = *body.Title
	}
	if body.Description != nil {
		task.Description = *body.Description
	}
	if body.Priority != nil {
		task.Priority = *body.Priority
	}
	if body.Status != nil {
		task.Status = *body.Status
	}
	task.AssignedTo = body.AssignedTo
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
	task.StartDate = startDate
	task.EndDate = endDate
	task.LabelIDs = body.LabelIDs

	if err := h.tasks.Update(r.Context(), task); err != nil {
		http.Error(w, "failed to update task", http.StatusInternalServerError)
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
