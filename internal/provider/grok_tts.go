package provider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

// GrokTTS implements iface.Synthesizer using Groq's API.
type GrokTTS struct {
	apiKey string
	client *http.Client
}

func NewGrokTTS(apiKey string) *GrokTTS {
	return &GrokTTS{apiKey: apiKey, client: &http.Client{}}
}

func (s *GrokTTS) Synthesize(text string) ([]byte, error) {
	slog.Info("tts.start", "provider", "groq", "text", text)

	payload := map[string]interface{}{
		"model":           "canopylabs/orpheus-arabic-saudi",
		"input":           text,
		"voice":           "noura",
		"response_format": "wav",
	}

	buf, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("tts: marshal request: %w", err)
	}

	req, err := http.NewRequest(
		http.MethodPost,
		"https://api.groq.com/openai/v1/audio/speech",
		bytes.NewReader(buf),
	)
	if err != nil {
		return nil, fmt.Errorf("tts: build request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		slog.Error("tts.request_failed", "error", err)
		return nil, fmt.Errorf("tts: request failed: %w", err)
	}
	defer resp.Body.Close()

	audioBytes, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		slog.Error("tts.non_200", "status", resp.StatusCode, "body", string(audioBytes))
		return nil, fmt.Errorf("tts: status %d: %s", resp.StatusCode, string(audioBytes))
	}

	slog.Info("tts.success", "audio_bytes", len(audioBytes))

	return audioBytes, nil
}