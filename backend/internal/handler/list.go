package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/pidanou/family-board/internal/service"
)

type ListHandler struct {
	lists *service.ListService
	hub   *Hub
}

func NewListHandler(lists *service.ListService, hub *Hub) *ListHandler {
	return &ListHandler{lists: lists, hub: hub}
}

func (h *ListHandler) Routes() http.Handler {
	r := chi.NewRouter()
	r.Get("/", h.listLists)
	r.Post("/", h.createList)
	r.Patch("/{listID}", h.renameList)
	r.Delete("/{listID}", h.deleteList)
	r.Get("/{listID}/items", h.listItems)
	r.Post("/{listID}/items", h.addItem)
	r.Patch("/{listID}/items/{itemID}", h.updateItem)
	r.Delete("/{listID}/items/checked", h.clearChecked) // must precede /{itemID}
	r.Delete("/{listID}/items/{itemID}", h.deleteItem)
	return r
}

func (h *ListHandler) listLists(w http.ResponseWriter, r *http.Request) {
	familyID := chi.URLParam(r, "familyID")
	lists, err := h.lists.ListsByFamily(r.Context(), familyID)
	if err != nil {
		http.Error(w, "failed to list lists", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(lists)
}

func (h *ListHandler) createList(w http.ResponseWriter, r *http.Request) {
	familyID := chi.URLParam(r, "familyID")
	var body struct{ Name string `json:"name"` }
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}
	list, err := h.lists.Create(r.Context(), familyID, body.Name)
	if err != nil {
		http.Error(w, "failed to create list", http.StatusInternalServerError)
		return
	}
	h.hub.Broadcast(familyID)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(list)
}

func (h *ListHandler) renameList(w http.ResponseWriter, r *http.Request) {
	familyID := chi.URLParam(r, "familyID")
	listID := chi.URLParam(r, "listID")
	var body struct{ Name string `json:"name"` }
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}
	if err := h.lists.Rename(r.Context(), listID, familyID, body.Name); err != nil {
		http.Error(w, "failed to rename list", http.StatusInternalServerError)
		return
	}
	h.hub.Broadcast(familyID)
	w.WriteHeader(http.StatusNoContent)
}

func (h *ListHandler) deleteList(w http.ResponseWriter, r *http.Request) {
	familyID := chi.URLParam(r, "familyID")
	listID := chi.URLParam(r, "listID")
	if err := h.lists.Delete(r.Context(), listID, familyID); err != nil {
		http.Error(w, "failed to delete list", http.StatusInternalServerError)
		return
	}
	h.hub.Broadcast(familyID)
	w.WriteHeader(http.StatusNoContent)
}

func (h *ListHandler) listItems(w http.ResponseWriter, r *http.Request) {
	listID := chi.URLParam(r, "listID")
	items, err := h.lists.ItemsByList(r.Context(), listID)
	if err != nil {
		http.Error(w, "failed to list items", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

func (h *ListHandler) addItem(w http.ResponseWriter, r *http.Request) {
	familyID := chi.URLParam(r, "familyID")
	listID := chi.URLParam(r, "listID")
	var body struct{ Name string `json:"name"` }
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}
	item, err := h.lists.AddItem(r.Context(), listID, familyID, body.Name)
	if err != nil {
		http.Error(w, "failed to add item", http.StatusInternalServerError)
		return
	}
	h.hub.Broadcast(familyID)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(item)
}

func (h *ListHandler) updateItem(w http.ResponseWriter, r *http.Request) {
	familyID := chi.URLParam(r, "familyID")
	listID := chi.URLParam(r, "listID")
	itemID := chi.URLParam(r, "itemID")
	var body struct {
		Name    string `json:"name"`
		Checked bool   `json:"checked"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	if err := h.lists.UpdateItem(r.Context(), itemID, listID, familyID, body.Name, body.Checked); err != nil {
		http.Error(w, "failed to update item", http.StatusInternalServerError)
		return
	}
	h.hub.Broadcast(familyID)
	w.WriteHeader(http.StatusNoContent)
}

func (h *ListHandler) deleteItem(w http.ResponseWriter, r *http.Request) {
	familyID := chi.URLParam(r, "familyID")
	listID := chi.URLParam(r, "listID")
	itemID := chi.URLParam(r, "itemID")
	if err := h.lists.DeleteItem(r.Context(), itemID, listID, familyID); err != nil {
		http.Error(w, "failed to delete item", http.StatusInternalServerError)
		return
	}
	h.hub.Broadcast(familyID)
	w.WriteHeader(http.StatusNoContent)
}

func (h *ListHandler) clearChecked(w http.ResponseWriter, r *http.Request) {
	familyID := chi.URLParam(r, "familyID")
	listID := chi.URLParam(r, "listID")
	if err := h.lists.ClearChecked(r.Context(), listID, familyID); err != nil {
		http.Error(w, "failed to clear checked items", http.StatusInternalServerError)
		return
	}
	h.hub.Broadcast(familyID)
	w.WriteHeader(http.StatusNoContent)
}
