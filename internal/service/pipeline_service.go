package service

import (
	"fmt"
	"log/slog"
	"time"

	"VA/internal/iface"
	"VA/internal/model"
)

// PipelineService orchestrates ASR -> clean -> Translate -> TTS.
type PipelineService struct {
	asr        iface.Transcriber
	translator iface.Translator
	tts        iface.Synthesizer
	repo       iface.SessionRepository
	cleaner    *TextCleaner
}

func NewPipelineService(asr iface.Transcriber, translator iface.Translator, tts iface.Synthesizer, repo iface.SessionRepository) *PipelineService {
	return &PipelineService{
		asr:        asr,
		translator: translator,
		tts:        tts,
		repo:       repo,
		cleaner:    NewTextCleaner(),
	}
}

func (p *PipelineService) Process(sessionID string, audio model.AudioChunk) ([]byte, error) {
	start := time.Now()
	slog.Info("pipeline.start", "session_id", sessionID)

	// Stage 1: ASR
	transcript, err := p.asr.Transcribe(audio)
	if err != nil {
		slog.Error("pipeline.asr_failed", "session_id", sessionID, "error", err)
		return nil, fmt.Errorf("pipeline: asr stage failed: %w", err)
	}
	if transcript.Text == "" {
		slog.Warn("pipeline.empty_transcript", "session_id", sessionID)
		return nil, fmt.Errorf("pipeline: empty transcript, nothing to translate")
	}

	// Stage 2: clean fillers before they hit the translator
	cleanText := p.cleaner.RemoveFillers(transcript.Text)
	slog.Info("pipeline.cleaned_text", "session_id", sessionID, "text", cleanText)

	// Stage 3: translate
	arabicText, err := p.translator.Translate(cleanText)
	if err != nil {
		slog.Error("pipeline.translation_failed", "session_id", sessionID, "error", err)
		return nil, fmt.Errorf("pipeline: translation stage failed: %w", err)
	}

	// Stage 4: synthesize
	audioOut, err := p.tts.Synthesize(arabicText)
	if err != nil {
		slog.Error("pipeline.tts_failed", "session_id", sessionID, "error", err)
		return nil, fmt.Errorf("pipeline: tts stage failed: %w", err)
	}

	result := model.TranslationResult{
		SourceText: transcript.Text,
		TargetText: arabicText,
		Tone:       "neutral",
	}
	if err := p.repo.Save(sessionID, result); err != nil {
		slog.Warn("pipeline.save_failed", "session_id", sessionID, "error", err)
	}

	slog.Info("pipeline.complete", "session_id", sessionID, "duration_ms", time.Since(start).Milliseconds())
	return audioOut, nil
}
