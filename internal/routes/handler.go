package routes

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"VA/internal/model"
	"VA/internal/service"

	"github.com/google/uuid"
)

// Handler exposes the pipeline over HTTP. 
type Handler struct {
	pipeline *service.PipelineService
}

func NewHandler(pipeline *service.PipelineService) *Handler {
	return &Handler{pipeline: pipeline}
}

// TranslateAudio: POST /translate  (multipart form field "audio")
// Returns Arabic speech as audio/mpeg bytes.
func (h *Handler) TranslateAudio(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	file, header, err := r.FormFile("audio")
	if err != nil {
		http.Error(w, "audio file required (multipart form field 'audio')", http.StatusBadRequest)
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "failed to read audio", http.StatusInternalServerError)
		return
	}

	sessionID := uuid.NewString()
	audio := model.AudioChunk{
		Data:     data,
		Format:   header.Filename,
		Filename: header.Filename,
	}

	slog.Info("http.request_received", "session_id", sessionID, "filename", header.Filename, "size_bytes", len(data))

	outAudio, err := h.pipeline.Process(sessionID, audio)
	if err != nil {
		slog.Error("http.pipeline_error", "session_id", sessionID, "error", err)
		http.Error(w, fmt.Sprintf("pipeline error: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "audio/mpeg")
	w.Header().Set("X-Session-Id", sessionID)
	w.Write(outAudio)
}
