package main

import (
	"log/slog"
	"net/http"
	"os"

	"VA/internal/provider"
	"VA/internal/repository"
	"VA/internal/routes"
	"VA/internal/service"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)
 
	groqKey := os.Getenv("GROQ_API_KEY")
	if groqKey == "" {
		slog.Error("startup.failed", "reason", "GROQ_API_KEY not set")
		os.Exit(1)
	}

	factory := provider.NewProviderFactory(groqKey)

	asr := factory.NewTranscriber("grok")
	translator := factory.NewTranslator("grok")
	tts := factory.NewSynthesizer("grok")

	repo := repository.NewMemorySessionRepository()

	pipeline := service.NewPipelineService(
		asr,
		translator,
		tts,
		repo,
	)

	handler := routes.NewHandler(pipeline)

	mux := http.NewServeMux()
	mux.HandleFunc("/translate", handler.TranslateAudio)
	mux.HandleFunc("/translate/stream", handler.TranslateStream)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	addr := ":8080"

	slog.Info("startup.server_started", "addr", addr)

	if err := http.ListenAndServe(addr, mux); err != nil {
		slog.Error("startup.server_crashed", "error", err)
		os.Exit(1)
	}
}
