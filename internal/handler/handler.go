package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/vibe-composer/internal/auth"
	"github.com/vibe-composer/internal/db"
	"github.com/vibe-composer/internal/storage"
	"github.com/vibe-composer/internal/worker"
)

// Handler holds dependencies for HTTP handlers.
type Handler struct {
	queries  *db.Queries
	gcs      *storage.GCS
	composer *worker.Composer
	bucket   string
}

// New creates a new Handler.
func New(queries *db.Queries, gcs *storage.GCS, composer *worker.Composer, bucket string) *Handler {
	return &Handler{
		queries:  queries,
		gcs:      gcs,
		composer: composer,
		bucket:   bucket,
	}
}

// GetMe returns the current authenticated user.
func (h *Handler) GetMe(w http.ResponseWriter, r *http.Request) {
	username := auth.UsernameFromContext(r.Context())
	writeJSON(w, http.StatusOK, map[string]string{"username": username})
}

// Compose handles music generation requests.
func (h *Handler) Compose(w http.ResponseWriter, r *http.Request) {
	username := auth.UsernameFromContext(r.Context())

	// Check for active generation
	active, err := h.queries.HasActiveGeneration(r.Context(), username)
	if err != nil {
		slog.Error("failed to check active generation", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	if active {
		writeJSON(w, http.StatusConflict, map[string]string{
			"error": "you already have a music generation in progress — please wait for it to finish",
		})
		return
	}

	// Parse multipart form (max 32MB)
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid form data"})
		return
	}

	textInput := r.FormValue("text")
	style := r.FormValue("style")
	if style == "" {
		style = "funny"
	}
	validStyles := map[string]bool{"funny": true, "harsh": true, "hiphop": true, "pansori": true}
	if !validStyles[style] {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "style must be 'funny', 'harsh', 'hiphop', or 'pansori'"})
		return
	}

	voice := r.FormValue("voice")
	if voice == "" {
		voice = "any"
	}
	validVoices := map[string]bool{"male": true, "female": true, "any": true}
	if !validVoices[voice] {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "voice must be 'male', 'female', or 'any'"})
		return
	}

	lyricType := r.FormValue("lyric_type")
	if lyricType == "" {
		lyricType = "arc"
	}
	validLyricTypes := map[string]bool{"arc": true, "immersion": true}
	if !validLyricTypes[lyricType] {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "lyric_type must be 'arc' or 'immersion'"})
		return
	}

	// Check for audio file
	var audioData []byte
	var audioMIME string
	file, header, err := r.FormFile("audio")
	if err == nil {
		defer file.Close()
		audioData, err = io.ReadAll(file)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "failed to read audio file"})
			return
		}
		audioMIME = header.Header.Get("Content-Type")
		if audioMIME == "" {
			audioMIME = "audio/webm"
		}
		slog.Info("received audio file", "size", len(audioData), "mime", audioMIME, "filename", header.Filename)
	}

	if textInput == "" && len(audioData) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "provide either text or audio input"})
		return
	}

	// Determine input type
	inputType := "text"
	if len(audioData) > 0 {
		inputType = "audio"
	}

	// Create composition record
	comp := &db.Composition{
		Username:    username,
		Status:      "pending",
		InputType:   inputType,
		MusicStyle:  style,
		VoiceGender: voice,
		LyricType:   lyricType,
	}
	if textInput != "" {
		comp.InputText = &textInput
	}

	// If audio, upload to GCS first
	var audioGCSPath string
	compID, err := h.queries.CreateComposition(r.Context(), comp)
	if err != nil {
		slog.Error("failed to create composition", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create composition"})
		return
	}

	if len(audioData) > 0 {
		audioGCSPath = fmt.Sprintf("inputs/%s/%s.webm", username, compID)
		if err := h.gcs.Upload(r.Context(), audioGCSPath, audioData, audioMIME); err != nil {
			slog.Error("failed to upload audio", "error", err)
			_ = h.queries.UpdateError(r.Context(), compID, "failed to upload audio")
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to upload audio"})
			return
		}
	}

	// Submit job to worker
	h.composer.Submit(worker.Job{
		CompositionID: compID,
		Username:      username,
		InputText:     textInput,
		AudioData:     audioData,
		AudioMIME:     audioMIME,
		MusicStyle:    style,
		VoiceGender:   voice,
		LyricType:     lyricType,
	})

	writeJSON(w, http.StatusAccepted, map[string]any{
		"id":      compID,
		"status":  "pending",
		"message": "your music is being composed! check back soon.",
	})
}

// ListCompositions returns all compositions for the current user.
func (h *Handler) ListCompositions(w http.ResponseWriter, r *http.Request) {
	username := auth.UsernameFromContext(r.Context())

	compositions, err := h.queries.ListCompositions(r.Context(), username)
	if err != nil {
		slog.Error("failed to list compositions", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}

	if compositions == nil {
		compositions = []*db.Composition{}
	}

	writeJSON(w, http.StatusOK, compositions)
}

// GetComposition returns a single composition by ID.
func (h *Handler) GetComposition(w http.ResponseWriter, r *http.Request) {
	username := auth.UsernameFromContext(r.Context())
	id := chi.URLParam(r, "id")

	comp, err := h.queries.GetComposition(r.Context(), id)
	if err != nil {
		slog.Error("failed to get composition", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	if comp == nil || comp.Username != username {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "composition not found"})
		return
	}

	writeJSON(w, http.StatusOK, comp)
}

// DownloadComposition streams the generated music file.
func (h *Handler) DownloadComposition(w http.ResponseWriter, r *http.Request) {
	username := auth.UsernameFromContext(r.Context())
	id := chi.URLParam(r, "id")

	comp, err := h.queries.GetComposition(r.Context(), id)
	if err != nil {
		slog.Error("failed to get composition", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	if comp == nil || comp.Username != username {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "composition not found"})
		return
	}
	if comp.Status != "done" || comp.ResultURL == nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "music not ready yet"})
		return
	}

	data, contentType, err := h.gcs.Download(r.Context(), *comp.ResultURL)
	if err != nil {
		slog.Error("failed to download from GCS", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to download"})
		return
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="vibe-%s.mp3"`, id[:8]))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(data)))
	w.Write(data)
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
