package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/pidanou/homeboard/internal/service"
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
	push  *service.PushService
}

func NewTaskHandler(tasks *service.TaskService, hub *Hub, push *service.PushService) *TaskHandler {
	return &TaskHandler{tasks: tasks, hub: hub, push: push}
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
		writeError(w, http.StatusInternalServerError, "list_tasks_failed")
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
		writeError(w, http.StatusBadRequest, "title_required")
		return
	}

	startDate, err := parseOptionalTime(body.StartDate)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_start_date")
		return
	}
	endDate, err := parseOptionalTime(body.EndDate)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_end_date")
		return
	}

	task, err := h.tasks.Create(r.Context(), familyID, userID, body.Title, body.Description, body.Important, body.AssignedTo, startDate, endDate, body.CategoryID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "create_task_failed")
		return
	}

	h.hub.Broadcast(familyID)
	go h.push.SendToFamily(context.WithoutCancel(r.Context()), familyID, "New task", task.Title)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(task)
}

func (h *TaskHandler) update(w http.ResponseWriter, r *http.Request) {
	familyID := chi.URLParam(r, "familyID")
	taskID := chi.URLParam(r, "taskID")

	task, err := h.tasks.GetByID(r.Context(), taskID, familyID)
	if err != nil {
		writeError(w, http.StatusNotFound, "task_not_found")
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
		writeError(w, http.StatusBadRequest, "invalid_request")
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
			writeError(w, http.StatusBadRequest, "invalid_start_date")
			return
		}
		task.StartDate = startDate
	}
	if body.EndDate != nil {
		endDate, err := parseOptionalTime(body.EndDate)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_end_date")
			return
		}
		task.EndDate = endDate
	}
	if body.CategoryID != nil {
		task.CategoryID = body.CategoryID
	}

	if err := h.tasks.Update(r.Context(), task); err != nil {
		writeError(w, http.StatusInternalServerError, "update_task_failed")
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
		writeError(w, http.StatusBadRequest, "ids_required")
		return
	}
	if err := h.tasks.Reorder(r.Context(), familyID, body.IDs); err != nil {
		writeError(w, http.StatusInternalServerError, "reorder_tasks_failed")
		return
	}
	h.hub.Broadcast(familyID)
	w.WriteHeader(http.StatusNoContent)
}

func (h *TaskHandler) delete(w http.ResponseWriter, r *http.Request) {
	familyID := chi.URLParam(r, "familyID")
	taskID := chi.URLParam(r, "taskID")

	if err := h.tasks.Delete(r.Context(), taskID, familyID); err != nil {
		writeError(w, http.StatusInternalServerError, "delete_task_failed")
		return
	}

	h.hub.Broadcast(familyID)
	w.WriteHeader(http.StatusNoContent)
}
