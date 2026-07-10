package model

// AudioChunk represents a raw audio input, either a full utterance
// file (assignment/demo mode) or a streamed segment (production mode).
type AudioChunk struct {
	Data     []byte
	Format   string // e.g. "wav", "mp3", "m4a"
	Filename string
}
