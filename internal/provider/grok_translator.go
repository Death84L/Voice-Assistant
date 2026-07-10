package provider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

// GrokTranslator implements iface.Translator using Groq's OpenAI-compatible
type GrokTranslator struct {
	apiKey string
	client *http.Client
}

func NewGrokTranslator(apiKey string) *GrokTranslator {
	return &GrokTranslator{
		apiKey: apiKey,
		client: &http.Client{},
	}
}

const translatorSystemPrompt = `You are an English-to-Arabic translator for a real-time voice assistant.

Rules:
- Translate into natural Modern Standard Arabic.
- Keep named entities and product names unchanged (Google Meet, Zoom, Asterisk, Microsoft Teams, etc.).
- Convert times, dates, and numbers into natural Arabic expressions (e.g. "5 PM" → "الساعة الخامسة مساءً").
- Remove unnecessary fillers such as "uh", "you know", "like" whenever appropriate.
- Preserve the original meaning and tone.
- Return ONLY the translated Arabic sentence.
- Do not explain your answer.
- Do not use quotes around the output.`

func (t *GrokTranslator) Translate(text string) (string, error) {
	slog.Info("translate.start", "provider", "groq", "input", text)

	payload := map[string]interface{}{
		"model": "llama-3.1-8b-instant",
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": translatorSystemPrompt,
			},
			{
				"role":    "user",
				"content": text,
			},
		},
		"temperature": 0.2,
	}

	buf, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("translate: marshal request: %w", err)
	}

	req, err := http.NewRequest(
		http.MethodPost,
		"https://api.groq.com/openai/v1/chat/completions",
		bytes.NewReader(buf),
	)
	if err != nil {
		return "", fmt.Errorf("translate: build request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+t.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := t.client.Do(req)
	if err != nil {
		slog.Error("translate.request_failed", "error", err)
		return "", fmt.Errorf("translate: request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		slog.Error("translate.non_200", "status", resp.StatusCode, "body", string(respBody))
		return "", fmt.Errorf("translate: status %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("translate: parse response: %w", err)
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("translate: empty response from model")
	}

	arabic := result.Choices[0].Message.Content

	slog.Info("translate.success", "output", arabic)

	return arabic, nil
}