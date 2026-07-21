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
	r.Patch("/locale", h.updateLocale)
	r.Patch("/time-format", h.updateTimeFormat)
	r.Patch("/reminder", h.updateReminder)
	r.Patch("/password", h.changePassword)
	r.Post("/avatar", h.uploadAvatar)
	r.Delete("/avatar", h.deleteAvatar)
	return r
}

func (h *ProfileHandler) get(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(ContextKeyUserID).(string)
	user, err := h.auth.GetProfile(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusNotFound, "not_found")
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
		writeError(w, http.StatusBadRequest, "name_required")
		return
	}
	user, err := h.auth.UpdateName(r.Context(), userID, strings.TrimSpace(body.Name))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "update_failed")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func (h *ProfileHandler) updateLocale(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(ContextKeyUserID).(string)
	var body struct {
		Locale string `json:"locale"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request")
		return
	}
	user, err := h.auth.SetLocale(r.Context(), userID, body.Locale)
	if err != nil {
		writeError(w, http.StatusBadRequest, "unsupported_locale")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func (h *ProfileHandler) updateTimeFormat(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(ContextKeyUserID).(string)
	var body struct {
		TimeFormat string `json:"time_format"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request")
		return
	}
	user, err := h.auth.SetTimeFormat(r.Context(), userID, body.TimeFormat)
	if err != nil {
		writeError(w, http.StatusBadRequest, "unsupported_time_format")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func (h *ProfileHandler) updateReminder(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(ContextKeyUserID).(string)
	var body struct {
		ReminderMinutesBefore *int `json:"reminder_minutes_before"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request")
		return
	}
	user, err := h.auth.SetReminderMinutesBefore(r.Context(), userID, body.ReminderMinutesBefore)
	if err != nil {
		writeError(w, http.StatusBadRequest, "unsupported_reminder_offset")
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
		writeError(w, http.StatusBadRequest, "invalid_request")
		return
	}
	if len(body.NewPassword) < 8 {
		writeError(w, http.StatusBadRequest, "password_too_short")
		return
	}
	if err := h.auth.ChangePassword(r.Context(), userID, body.CurrentPassword, body.NewPassword); err != nil {
		writeError(w, http.StatusBadRequest, codeFor(err, "invalid_request"))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *ProfileHandler) uploadAvatar(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(ContextKeyUserID).(string)

	// 6 MB limit (5 MB file + form overhead)
	r.Body = http.MaxBytesReader(w, r.Body, 6<<20)
	if err := r.ParseMultipartForm(6 << 20); err != nil {
		writeError(w, http.StatusRequestEntityTooLarge, "file_too_large")
		return
	}

	file, _, err := r.FormFile("avatar")
	if err != nil {
		writeError(w, http.StatusBadRequest, "image_required")
		return
	}
	defer file.Close()

	sniff := make([]byte, 512)
	n, _ := file.Read(sniff)
	ct := http.DetectContentType(sniff[:n])
	file.Seek(0, io.SeekStart)
	if ct != "image/jpeg" && ct != "image/png" && ct != "image/webp" {
		writeError(w, http.StatusBadRequest, "unsupported_image_type")
		return
	}

	ext := map[string]string{
		"image/jpeg": ".jpg",
		"image/png":  ".png",
		"image/webp": ".webp",
	}[ct]

	if err := os.MkdirAll(filepath.Join(h.uploadDir, "avatars"), 0755); err != nil {
		writeError(w, http.StatusInternalServerError, "storage_error")
		return
	}

	filename := uuid.NewString() + ext
	dst, err := os.Create(filepath.Join(h.uploadDir, "avatars", filename))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "storage_error")
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		writeError(w, http.StatusInternalServerError, "storage_error")
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
		writeError(w, http.StatusInternalServerError, "update_failed")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"avatar_url": avatarURL})
}

func (h *ProfileHandler) deleteAvatar(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(ContextKeyUserID).(string)

	user, err := h.auth.GetProfile(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusNotFound, "not_found")
		return
	}
	if user.AvatarURL != nil {
		h.deleteAvatarFile(*user.AvatarURL)
	}

	if err := h.auth.SetAvatar(r.Context(), userID, nil); err != nil {
		writeError(w, http.StatusInternalServerError, "update_failed")
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
