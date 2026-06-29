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

type HouseholdHandler struct {
	families  *service.HouseholdService
	uploadDir string
	baseURL   string
}

func NewHouseholdHandler(families *service.HouseholdService, uploadDir, baseURL string) *HouseholdHandler {
	return &HouseholdHandler{families: families, uploadDir: uploadDir, baseURL: baseURL}
}

func (h *HouseholdHandler) Routes() http.Handler {
	r := chi.NewRouter()
	r.Get("/", h.list)
	r.Post("/", h.create)
	r.Get("/{familyID}", h.get)
	r.Patch("/{familyID}", h.updateName)
	r.Get("/{familyID}/members", h.members)
	r.Post("/{familyID}/members/virtual", h.createVirtual)
	r.Delete("/{familyID}/members/{memberID}", h.removeMember)
	r.Put("/{familyID}/members/{memberID}/role", h.updateRole)
	r.Delete("/{familyID}/members/virtual/{memberID}", h.deleteVirtual)
	r.Post("/{familyID}/members/virtual/{memberID}/link", h.linkVirtual)
	r.Get("/{familyID}/members/virtual/unlinked", h.unlinkedVirtual)
	r.Get("/{familyID}/photo", h.servePhoto)
	r.Post("/{familyID}/photo", h.uploadPhoto)
	r.Delete("/{familyID}/photo", h.deletePhoto)
	r.Get("/{familyID}/wallpaper", h.serveWallpaper)
	r.Post("/{familyID}/wallpaper", h.uploadWallpaper)
	r.Delete("/{familyID}/wallpaper", h.deleteWallpaper)
	return r
}

func (h *HouseholdHandler) list(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(ContextKeyUserID).(string)

	families, err := h.families.ListForUser(r.Context(), userID)
	if err != nil {
		http.Error(w, "failed to list families", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(families)
}

func (h *HouseholdHandler) create(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(ContextKeyUserID).(string)

	if os.Getenv("ALLOW_MULTI_HOUSEHOLD") != "true" {
		exists, err := h.families.Exists(r.Context())
		if err != nil {
			http.Error(w, "failed to check households", http.StatusInternalServerError)
			return
		}
		if exists {
			http.Error(w, "only one household allowed in single-family mode", http.StatusForbidden)
			return
		}
	}

	var body struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	family, err := h.families.Create(r.Context(), body.Name, userID)
	if err != nil {
		http.Error(w, "failed to create family", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(family)
}

func (h *HouseholdHandler) get(w http.ResponseWriter, r *http.Request) {
	familyID := chi.URLParam(r, "familyID")
	if err := requireMember(r, familyID, h.families); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	family, err := h.families.GetByID(r.Context(), familyID)
	if err != nil {
		http.Error(w, "family not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(family)
}

func (h *HouseholdHandler) members(w http.ResponseWriter, r *http.Request) {
	familyID := chi.URLParam(r, "familyID")
	if err := requireMember(r, familyID, h.families); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	members, err := h.families.GetMembers(r.Context(), familyID)
	if err != nil {
		http.Error(w, "failed to get members", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(members)
}

func (h *HouseholdHandler) createVirtual(w http.ResponseWriter, r *http.Request) {
	familyID := chi.URLParam(r, "familyID")
	callerID, ok := r.Context().Value(ContextKeyUserID).(string)
	if !ok || callerID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	var body struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name == "" {
		http.Error(w, "name required", http.StatusBadRequest)
		return
	}
	m, err := h.families.CreateVirtualMember(r.Context(), familyID, body.Name, callerID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(m)
}

func (h *HouseholdHandler) deleteVirtual(w http.ResponseWriter, r *http.Request) {
	familyID := chi.URLParam(r, "familyID")
	memberID := chi.URLParam(r, "memberID")
	callerID, ok := r.Context().Value(ContextKeyUserID).(string)
	if !ok || callerID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if err := h.families.DeleteVirtualMember(r.Context(), memberID, familyID, callerID); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *HouseholdHandler) updateRole(w http.ResponseWriter, r *http.Request) {
	familyID := chi.URLParam(r, "familyID")
	memberID := chi.URLParam(r, "memberID")
	callerID, ok := r.Context().Value(ContextKeyUserID).(string)
	if !ok || callerID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	var body struct {
		Role string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	if err := h.families.UpdateMemberRole(r.Context(), memberID, familyID, body.Role, callerID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *HouseholdHandler) unlinkedVirtual(w http.ResponseWriter, r *http.Request) {
	familyID := chi.URLParam(r, "familyID")
	if err := requireMember(r, familyID, h.families); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	members, err := h.families.GetUnlinkedVirtualMembers(r.Context(), familyID)
	if err != nil {
		http.Error(w, "failed to get virtual members", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(members)
}

func (h *HouseholdHandler) removeMember(w http.ResponseWriter, r *http.Request) {
	familyID := chi.URLParam(r, "familyID")
	memberID := chi.URLParam(r, "memberID")
	callerID, ok := r.Context().Value(ContextKeyUserID).(string)
	if !ok || callerID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if err := h.families.RemoveMember(r.Context(), memberID, familyID, callerID); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *HouseholdHandler) updateName(w http.ResponseWriter, r *http.Request) {
	familyID := chi.URLParam(r, "familyID")
	callerID, ok := r.Context().Value(ContextKeyUserID).(string)
	if !ok || callerID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	var body struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name == "" {
		http.Error(w, "name required", http.StatusBadRequest)
		return
	}
	if err := h.families.UpdateName(r.Context(), familyID, body.Name, callerID); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *HouseholdHandler) linkVirtual(w http.ResponseWriter, r *http.Request) {
	familyID := chi.URLParam(r, "familyID")
	memberID := chi.URLParam(r, "memberID")
	userID, ok := r.Context().Value(ContextKeyUserID).(string)
	if !ok || userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if err := h.families.LinkVirtualMember(r.Context(), memberID, familyID, userID); err != nil {
		http.Error(w, "failed to link virtual member", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *HouseholdHandler) servePhoto(w http.ResponseWriter, r *http.Request) {
	h.serveHouseholdImage(w, r, "photo")
}

func (h *HouseholdHandler) serveWallpaper(w http.ResponseWriter, r *http.Request) {
	h.serveHouseholdImage(w, r, "wallpaper")
}

func (h *HouseholdHandler) serveHouseholdImage(w http.ResponseWriter, r *http.Request, kind string) {
	familyID := chi.URLParam(r, "familyID")
	if err := requireMember(r, familyID, h.families); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	family, err := h.families.GetByID(r.Context(), familyID)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	var storedURL *string
	if kind == "photo" {
		storedURL = family.PhotoURL
	} else {
		storedURL = family.WallpaperURL
	}
	if storedURL == nil {
		http.NotFound(w, r)
		return
	}
	// Strip URL prefix to get path relative to uploadDir
	prefix := "/api/v1/uploads/"
	rel := strings.TrimPrefix(*storedURL, h.baseURL+prefix)
	rel = strings.TrimPrefix(rel, prefix)
	http.ServeFile(w, r, filepath.Join(h.uploadDir, rel))
}

func (h *HouseholdHandler) uploadPhoto(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "familyID")
	h.uploadHouseholdImage(w, r, "photo", "household/"+id+"/photos")
}

func (h *HouseholdHandler) deletePhoto(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "familyID")
	h.deleteHouseholdImage(w, r, "photo", "household/"+id+"/photos")
}

func (h *HouseholdHandler) uploadWallpaper(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "familyID")
	h.uploadHouseholdImage(w, r, "wallpaper", "household/"+id+"/wallpapers")
}

func (h *HouseholdHandler) deleteWallpaper(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "familyID")
	h.deleteHouseholdImage(w, r, "wallpaper", "household/"+id+"/wallpapers")
}

func (h *HouseholdHandler) uploadHouseholdImage(w http.ResponseWriter, r *http.Request, kind, subdir string) {
	familyID := chi.URLParam(r, "familyID")
	callerID, ok := r.Context().Value(ContextKeyUserID).(string)
	if !ok || callerID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 10<<20)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "file too large", http.StatusRequestEntityTooLarge)
		return
	}

	file, _, err := r.FormFile(kind)
	if err != nil {
		http.Error(w, kind+" file required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	sniff := make([]byte, 512)
	n, _ := file.Read(sniff)
	ct := http.DetectContentType(sniff[:n])
	file.Seek(0, io.SeekStart)
	extMap := map[string]string{"image/jpeg": ".jpg", "image/png": ".png", "image/webp": ".webp"}
	ext, ok := extMap[ct]
	if !ok {
		http.Error(w, "unsupported image type", http.StatusBadRequest)
		return
	}

	if err := os.MkdirAll(filepath.Join(h.uploadDir, subdir), 0755); err != nil {
		http.Error(w, "storage error", http.StatusInternalServerError)
		return
	}

	filename := uuid.NewString() + ext
	dst, err := os.Create(filepath.Join(h.uploadDir, subdir, filename))
	if err != nil {
		http.Error(w, "storage error", http.StatusInternalServerError)
		return
	}
	defer dst.Close()
	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, "storage error", http.StatusInternalServerError)
		return
	}

	relPath := fmt.Sprintf("/api/v1/uploads/%s/%s", subdir, filename)
	var imageURL string
	if h.baseURL != "" {
		imageURL = h.baseURL + relPath
	} else {
		imageURL = relPath
	}

	// Delete old cropped file
	old, _ := h.families.GetByID(r.Context(), familyID)
	if old != nil {
		var oldURL *string
		if kind == "photo" {
			oldURL = old.PhotoURL
		} else {
			oldURL = old.WallpaperURL
		}
		if oldURL != nil {
			h.deleteLocalFile(*oldURL, subdir)
		}
	}

	var setErr error
	if kind == "photo" {
		setErr = h.families.SetPhoto(r.Context(), familyID, callerID, &imageURL)
	} else {
		setErr = h.families.SetWallpaper(r.Context(), familyID, callerID, &imageURL)
	}
	if setErr != nil {
		http.Error(w, setErr.Error(), http.StatusForbidden)
		return
	}

	// Optionally save original if provided
	originalKind := kind + "_original"
	origFile, _, origErr := r.FormFile(originalKind)
	originalURL := ""
	if origErr == nil {
		defer origFile.Close()

		sniff2 := make([]byte, 512)
		n2, _ := origFile.Read(sniff2)
		ct2 := http.DetectContentType(sniff2[:n2])
		origFile.Seek(0, io.SeekStart)
		origExt, ok2 := extMap[ct2]
		if ok2 {
			origSubdir := subdir + "/originals"
			if err := os.MkdirAll(filepath.Join(h.uploadDir, origSubdir), 0755); err == nil {
				origFilename := uuid.NewString() + origExt
				if dst2, err := os.Create(filepath.Join(h.uploadDir, origSubdir, origFilename)); err == nil {
					io.Copy(dst2, origFile)
					dst2.Close()
					originalURL = fmt.Sprintf("/api/v1/uploads/%s/%s", origSubdir, origFilename)
					if h.baseURL != "" {
						originalURL = h.baseURL + fmt.Sprintf("/api/v1/uploads/%s/%s", origSubdir, origFilename)
					}
					// Delete old original
					if old != nil {
						var oldOrigURL *string
						if kind == "photo" {
							oldOrigURL = old.PhotoOriginalURL
						} else {
							oldOrigURL = old.WallpaperOriginalURL
						}
						if oldOrigURL != nil {
							h.deleteLocalFile(*oldOrigURL, origSubdir)
						}
					}
					if kind == "photo" {
						h.families.SetPhotoOriginal(r.Context(), familyID, callerID, &originalURL)
					} else {
						h.families.SetWallpaperOriginal(r.Context(), familyID, callerID, &originalURL)
					}
				}
			}
		}
	}

	resp := map[string]string{"url": imageURL}
	if originalURL != "" {
		resp["original_url"] = originalURL
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *HouseholdHandler) deleteHouseholdImage(w http.ResponseWriter, r *http.Request, kind, subdir string) {
	familyID := chi.URLParam(r, "familyID")
	callerID, ok := r.Context().Value(ContextKeyUserID).(string)
	if !ok || callerID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	old, err := h.families.GetByID(r.Context(), familyID)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	var oldURL *string
	if kind == "photo" {
		oldURL = old.PhotoURL
	} else {
		oldURL = old.WallpaperURL
	}
	if oldURL != nil {
		h.deleteLocalFile(*oldURL, subdir)
	}

	var setErr error
	if kind == "photo" {
		setErr = h.families.SetPhoto(r.Context(), familyID, callerID, nil)
	} else {
		setErr = h.families.SetWallpaper(r.Context(), familyID, callerID, nil)
	}
	if setErr != nil {
		http.Error(w, setErr.Error(), http.StatusForbidden)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *HouseholdHandler) deleteLocalFile(url, subdir string) {
	relPrefix := fmt.Sprintf("/api/v1/uploads/%s/", subdir)
	var filename string
	if strings.HasPrefix(url, relPrefix) {
		filename = strings.TrimPrefix(url, relPrefix)
	} else if h.baseURL != "" && strings.HasPrefix(url, h.baseURL+relPrefix) {
		filename = strings.TrimPrefix(url, h.baseURL+relPrefix)
	} else {
		return
	}
	os.Remove(filepath.Join(h.uploadDir, subdir, filename)) //nolint:errcheck
}
