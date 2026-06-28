package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/pidanou/homeboard/internal/service"
)

type ProfileHandler struct {
	auth      *service.AuthService
	uploadDir string
	baseURL   string
}

func NewProfileHandler(auth *service.AuthService, uploadDir, baseURL string) *ProfileHandler {
	return &ProfileHandler{auth: auth, uploadDir: uploadDir, baseURL: baseURL}
}

func (h *ProfileHandler) Routes() http.Handler {
	r := chi.NewRouter()
	r.Get("/", h.get)
	r.Patch("/", h.updateName)
	r.Patch("/password", h.changePassword)
	r.Post("/avatar", h.uploadAvatar)
	r.Delete("/avatar", h.deleteAvatar)
	return r
}

func (h *ProfileHandler) get(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(ContextKeyUserID).(string)
	user, err := h.auth.GetProfile(r.Context(), userID)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func (h *ProfileHandler) updateName(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(ContextKeyUserID).(string)
	var body struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || strings.TrimSpace(body.Name) == "" {
		http.Error(w, "name required", http.StatusBadRequest)
		return
	}
	user, err := h.auth.UpdateName(r.Context(), userID, strings.TrimSpace(body.Name))
	if err != nil {
		http.Error(w, "update failed", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func (h *ProfileHandler) changePassword(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(ContextKeyUserID).(string)
	var body struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	if len(body.NewPassword) < 8 {
		http.Error(w, "password must be at least 8 characters", http.StatusBadRequest)
		return
	}
	if err := h.auth.ChangePassword(r.Context(), userID, body.CurrentPassword, body.NewPassword); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *ProfileHandler) uploadAvatar(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(ContextKeyUserID).(string)

	// 6 MB limit (5 MB file + form overhead)
	r.Body = http.MaxBytesReader(w, r.Body, 6<<20)
	if err := r.ParseMultipartForm(6 << 20); err != nil {
		http.Error(w, "file too large", http.StatusRequestEntityTooLarge)
		return
	}

	file, _, err := r.FormFile("avatar")
	if err != nil {
		http.Error(w, "avatar file required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	sniff := make([]byte, 512)
	n, _ := file.Read(sniff)
	ct := http.DetectContentType(sniff[:n])
	file.Seek(0, io.SeekStart)
	if ct != "image/jpeg" && ct != "image/png" && ct != "image/webp" {
		http.Error(w, "unsupported image type", http.StatusBadRequest)
		return
	}

	ext := map[string]string{
		"image/jpeg": ".jpg",
		"image/png":  ".png",
		"image/webp": ".webp",
	}[ct]

	if err := os.MkdirAll(filepath.Join(h.uploadDir, "avatars"), 0755); err != nil {
		http.Error(w, "storage error", http.StatusInternalServerError)
		return
	}

	filename := uuid.NewString() + ext
	dst, err := os.Create(filepath.Join(h.uploadDir, "avatars", filename))
	if err != nil {
		http.Error(w, "storage error", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, "storage error", http.StatusInternalServerError)
		return
	}

	var avatarURL string
	if h.baseURL != "" {
		avatarURL = fmt.Sprintf("%s/api/v1/uploads/avatars/%s", h.baseURL, filename)
	} else {
		avatarURL = fmt.Sprintf("/api/v1/uploads/avatars/%s", filename)
	}

	// Delete old avatar file if it was locally stored
	oldUser, _ := h.auth.GetProfile(r.Context(), userID)
	if oldUser != nil && oldUser.AvatarURL != nil {
		h.deleteAvatarFile(*oldUser.AvatarURL)
	}

	if err := h.auth.SetAvatar(r.Context(), userID, &avatarURL); err != nil {
		http.Error(w, "update failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"avatar_url": avatarURL})
}

func (h *ProfileHandler) deleteAvatar(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(ContextKeyUserID).(string)

	user, err := h.auth.GetProfile(r.Context(), userID)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if user.AvatarURL != nil {
		h.deleteAvatarFile(*user.AvatarURL)
	}

	if err := h.auth.SetAvatar(r.Context(), userID, nil); err != nil {
		http.Error(w, "update failed", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// deleteAvatarFile removes a locally-stored avatar. Silently ignores errors (file may not exist).
func (h *ProfileHandler) deleteAvatarFile(avatarURL string) {
	const relPrefix = "/api/v1/uploads/avatars/"
	var filename string
	if strings.HasPrefix(avatarURL, relPrefix) {
		filename = strings.TrimPrefix(avatarURL, relPrefix)
	} else if h.baseURL != "" && strings.HasPrefix(avatarURL, h.baseURL+relPrefix) {
		filename = strings.TrimPrefix(avatarURL, h.baseURL+relPrefix)
	} else {
		return // external URL, don't touch
	}
	os.Remove(filepath.Join(h.uploadDir, "avatars", filename)) //nolint:errcheck
}
