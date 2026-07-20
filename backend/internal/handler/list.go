package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/pidanou/homeboard/internal/service"
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
	r.Patch("/reorder", h.reorderLists) // static route, must precede /{listID}
	r.Patch("/{listID}", h.renameList)
	r.Delete("/{listID}", h.deleteList)
	r.Get("/{listID}/items", h.listItems)
	r.Post("/{listID}/items", h.addItem)
	r.Patch("/{listID}/items/reorder", h.reorderItems) // static route, must precede /{itemID}
	r.Patch("/{listID}/items/{itemID}", h.updateItem)
	r.Delete("/{listID}/items/checked", h.clearChecked) // must precede /{itemID}
	r.Delete("/{listID}/items/{itemID}", h.deleteItem)
	return r
}

func (h *ListHandler) listLists(w http.ResponseWriter, r *http.Request) {
	familyID := chi.URLParam(r, "familyID")
	lists, err := h.lists.ListsByFamily(r.Context(), familyID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "list_lists_failed")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(lists)
}

func (h *ListHandler) createList(w http.ResponseWriter, r *http.Request) {
	familyID := chi.URLParam(r, "familyID")
	var body struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name == "" {
		writeError(w, http.StatusBadRequest, "name_required")
		return
	}
	list, err := h.lists.Create(r.Context(), familyID, body.Name)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "create_list_failed")
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
	var body struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name == "" {
		writeError(w, http.StatusBadRequest, "name_required")
		return
	}
	if err := h.lists.Rename(r.Context(), listID, familyID, body.Name); err != nil {
		writeError(w, http.StatusInternalServerError, "rename_list_failed")
		return
	}
	h.hub.Broadcast(familyID)
	w.WriteHeader(http.StatusNoContent)
}

func (h *ListHandler) reorderLists(w http.ResponseWriter, r *http.Request) {
	familyID := chi.URLParam(r, "familyID")
	var body struct {
		ListIDs []string `json:"list_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || len(body.ListIDs) == 0 {
		writeError(w, http.StatusBadRequest, "ids_required")
		return
	}
	if err := h.lists.Reorder(r.Context(), familyID, body.ListIDs); err != nil {
		writeError(w, http.StatusInternalServerError, "reorder_lists_failed")
		return
	}
	h.hub.Broadcast(familyID)
	w.WriteHeader(http.StatusNoContent)
}

func (h *ListHandler) deleteList(w http.ResponseWriter, r *http.Request) {
	familyID := chi.URLParam(r, "familyID")
	listID := chi.URLParam(r, "listID")
	if err := h.lists.Delete(r.Context(), listID, familyID); err != nil {
		writeError(w, http.StatusInternalServerError, "delete_list_failed")
		return
	}
	h.hub.Broadcast(familyID)
	w.WriteHeader(http.StatusNoContent)
}

func (h *ListHandler) listItems(w http.ResponseWriter, r *http.Request) {
	listID := chi.URLParam(r, "listID")
	items, err := h.lists.ItemsByList(r.Context(), listID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "list_items_failed")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

func (h *ListHandler) addItem(w http.ResponseWriter, r *http.Request) {
	familyID := chi.URLParam(r, "familyID")
	listID := chi.URLParam(r, "listID")
	var body struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name == "" {
		writeError(w, http.StatusBadRequest, "name_required")
		return
	}
	item, err := h.lists.AddItem(r.Context(), listID, familyID, body.Name)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "add_item_failed")
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
		Name     string  `json:"name"`
		Checked  bool    `json:"checked"`
		Category *string `json:"category"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request")
		return
	}
	if err := h.lists.UpdateItem(r.Context(), itemID, listID, familyID, body.Name, body.Checked, body.Category); err != nil {
		writeError(w, http.StatusInternalServerError, "update_item_failed")
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
		writeError(w, http.StatusInternalServerError, "delete_item_failed")
		return
	}
	h.hub.Broadcast(familyID)
	w.WriteHeader(http.StatusNoContent)
}

func (h *ListHandler) clearChecked(w http.ResponseWriter, r *http.Request) {
	familyID := chi.URLParam(r, "familyID")
	listID := chi.URLParam(r, "listID")
	if err := h.lists.ClearChecked(r.Context(), listID, familyID); err != nil {
		writeError(w, http.StatusInternalServerError, "clear_checked_items_failed")
		return
	}
	h.hub.Broadcast(familyID)
	w.WriteHeader(http.StatusNoContent)
}

func (h *ListHandler) reorderItems(w http.ResponseWriter, r *http.Request) {
	familyID := chi.URLParam(r, "familyID")
	listID := chi.URLParam(r, "listID")
	var body struct {
		IDs []string `json:"ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || len(body.IDs) == 0 {
		writeError(w, http.StatusBadRequest, "ids_required")
		return
	}
	if err := h.lists.ReorderItems(r.Context(), listID, familyID, body.IDs); err != nil {
		writeError(w, http.StatusInternalServerError, "reorder_items_failed")
		return
	}
	h.hub.Broadcast(familyID)
	w.WriteHeader(http.StatusNoContent)
}
