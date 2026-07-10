package service_test

import (
	"errors"
	"testing"

	"VA/internal/model"
	"VA/internal/repository"
	"VA/internal/service"
)

// --- mocks implementing the same interfaces as real providers ---

type mockTranscriber struct {
	text string
	err  error
}

func (m *mockTranscriber) Transcribe(audio model.AudioChunk) (model.Transcript, error) {
	if m.err != nil {
		return model.Transcript{}, m.err
	}
	return model.Transcript{Text: m.text, Confidence: 1.0, Language: "en"}, nil
}

type mockTranslator struct {
	output string
	err    error
}

func (m *mockTranslator) Translate(text string) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return m.output, nil
}

type mockSynthesizer struct {
	audio []byte
	err   error
}

func (m *mockSynthesizer) Synthesize(text string) ([]byte, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.audio, nil
}

// --- tests ---

func TestPipelineService_Process_Success(t *testing.T) {
	asr := &mockTranscriber{text: "Can we move the meeting to 5 PM tomorrow?"}
	translator := &mockTranslator{output: "क्या हम मीटिंग को कल शाम 5 बजे कर सकते हैं?"}
	tts := &mockSynthesizer{audio: []byte("fake-audio-bytes")}
	repo := repository.NewMemorySessionRepository()

	pipeline := service.NewPipelineService(asr, translator, tts, repo)

	out, err := pipeline.Process("session-1", model.AudioChunk{Data: []byte("fake-wav"), Filename: "sample.wav"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if string(out) != "fake-audio-bytes" {
		t.Errorf("expected synthesized audio, got %s", out)
	}

	saved, err := repo.Get("session-1")
	if err != nil {
		t.Fatalf("expected saved session, got error: %v", err)
	}
	if saved.TargetText != "क्या हम मीटिंग को कल शाम 5 बजे कर सकते हैं?" {
		t.Errorf("unexpected saved translation: %s", saved.TargetText)
	}
}

func TestPipelineService_Process_EmptyTranscript(t *testing.T) {
	asr := &mockTranscriber{text: ""}
	translator := &mockTranslator{}
	tts := &mockSynthesizer{}
	repo := repository.NewMemorySessionRepository()

	pipeline := service.NewPipelineService(asr, translator, tts, repo)

	_, err := pipeline.Process("session-2", model.AudioChunk{Data: []byte("fake"), Filename: "sample.wav"})
	if err == nil {
		t.Fatal("expected error for empty transcript, got nil")
	}
}

func TestPipelineService_Process_ASRFailure(t *testing.T) {
	asr := &mockTranscriber{err: errors.New("asr provider down")}
	translator := &mockTranslator{}
	tts := &mockSynthesizer{}
	repo := repository.NewMemorySessionRepository()

	pipeline := service.NewPipelineService(asr, translator, tts, repo)

	_, err := pipeline.Process("session-3", model.AudioChunk{Data: []byte("fake"), Filename: "sample.wav"})
	if err == nil {
		t.Fatal("expected error when ASR fails, got nil")
	}
}

func TestPipelineService_Process_TranslationFailure(t *testing.T) {
	asr := &mockTranscriber{text: "hello"}
	translator := &mockTranslator{err: errors.New("translation provider down")}
	tts := &mockSynthesizer{}
	repo := repository.NewMemorySessionRepository()

	pipeline := service.NewPipelineService(asr, translator, tts, repo)

	_, err := pipeline.Process("session-4", model.AudioChunk{Data: []byte("fake"), Filename: "sample.wav"})
	if err == nil {
		t.Fatal("expected error when translation fails, got nil")
	}
}
