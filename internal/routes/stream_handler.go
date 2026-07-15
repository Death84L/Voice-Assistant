package routes

import (
	"fmt"
	"log/slog"
	"net/http"

	"VA/internal/model"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (h *Handler) TranslateStream(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("ws.upgrade_failed", "error", err)
		return
	}
	defer conn.Close()

	sessionID := uuid.NewString()
	seq := 0

	slog.Info("ws.session_started", "session_id", sessionID)

	for {
		msgType, data, err := conn.ReadMessage()
		if err != nil {
			slog.Info("ws.session_ended", "session_id", sessionID, "chunks_processed", seq, "reason", err)
			return
		}

		// Only care about binary audio chunks.
		if msgType != websocket.BinaryMessage {
			continue
		}
		if len(data) == 0 {
			continue
		}

		chunk := model.AudioChunk{
			Data:     data,
			Format:   "webm",
			Filename: fmt.Sprintf("chunk-%d.webm", seq),
		}
		seq++

		slog.Info("ws.chunk_received", "session_id", sessionID, "seq", seq, "size_bytes", len(data))

		outAudio, err := h.pipeline.Process(sessionID, chunk)
		if err != nil {
			// Skip a bad/silent chunk rather than killing the whole session 
			slog.Warn("ws.chunk_failed", "session_id", sessionID, "seq", seq, "error", err)
			continue
		}

		if err := conn.WriteMessage(websocket.BinaryMessage, outAudio); err != nil {
			slog.Error("ws.write_failed", "session_id", sessionID, "seq", seq, "error", err)
			return
		}
	}
}
