package routes

import (
	"fmt"
	"log/slog"
	"net/http"

	"VA/internal/model"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// upgrader upgrades a plain HTTP connection to a WebSocket connection.
// CheckOrigin is left permissive here since this is an assignment/demo
// running locally; lock this down for anything beyond that.
var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// TranslateStream: WS /translate/stream
//
// Protocol:
//   - Client opens the WebSocket connection (one session per connection).
//   - Client sends one BinaryMessage per audio chunk. Each chunk MUST be a
//     complete, self-contained audio file (e.g. a 3s WAV/WebM/OGG clip),
//     not a raw slice of a continuous stream — the ASR provider needs a
//     decodable file per call.
//   - Server responds with one BinaryMessage per chunk: the translated
//     speech audio for that chunk, in the same order it was received.
//   - Connection closes -> session ends. No reconnect/resume support.
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

		// Ignore text/control frames; we only care about binary audio chunks.
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
			// Skip a bad/silent chunk rather than killing the whole session -
			// e.g. a chunk with no speech will fail at the ASR stage with an
			// empty transcript, which is expected during silence.
			slog.Warn("ws.chunk_failed", "session_id", sessionID, "seq", seq, "error", err)
			continue
		}

		if err := conn.WriteMessage(websocket.BinaryMessage, outAudio); err != nil {
			slog.Error("ws.write_failed", "session_id", sessionID, "seq", seq, "error", err)
			return
		}
	}
}