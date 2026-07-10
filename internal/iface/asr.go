package iface

import "VA/internal/model"

// Transcriber converts audio into English text.
type Transcriber interface {
	Transcribe(audio model.AudioChunk) (model.Transcript, error)
}
