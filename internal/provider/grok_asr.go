package provider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"

	"VA/internal/model"
)

// GrokASR implements iface.Transcriber using Groq's API.
type GrokASR struct {
	apiKey string
	client *http.Client
}

func NewGrokASR(apiKey string) *GrokASR {
	return &GrokASR{apiKey: apiKey, client: &http.Client{}}
}

func (g *GrokASR) Transcribe(audio model.AudioChunk) (model.Transcript, error) {
	slog.Info("asr.start", "provider", "grok", "filename", audio.Filename, "size_bytes", len(audio.Data))

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Groq uses OpenAI-compatible audio transcription requests.
	_ = writer.WriteField("model", "whisper-large-v3")
	_ = writer.WriteField("language", "en")

	part, err := writer.CreateFormFile("file", audio.Filename)
	if err != nil {
		return model.Transcript{}, fmt.Errorf("asr: create form file: %w", err)
	}
	if _, err := part.Write(audio.Data); err != nil {
		return model.Transcript{}, fmt.Errorf("asr: write audio data: %w", err)
	}
	writer.Close()

	req, err := http.NewRequest(http.MethodPost, "https://api.groq.com/openai/v1/audio/transcriptions", body)
	if err != nil {
		return model.Transcript{}, fmt.Errorf("asr: build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+g.apiKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := g.client.Do(req)
	if err != nil {
		slog.Error("asr.request_failed", "error", err)
		return model.Transcript{}, fmt.Errorf("asr: request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		slog.Error("asr.non_200", "status", resp.StatusCode, "body", string(respBody))
		return model.Transcript{}, fmt.Errorf("asr: status %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return model.Transcript{}, fmt.Errorf("asr: parse response: %w", err)
	}

	slog.Info("asr.success", "text", result.Text)
	return model.Transcript{Text: result.Text, Confidence: 1.0, Language: "en"}, nil
}