package repository

import (
	"fmt"
	"sync"

	"VA/internal/model"
)

// MemorySessionRepository is a thread-safe in-memory implementation of SessionRepository.
type MemorySessionRepository struct {
	mu   sync.RWMutex
	data map[string]model.TranslationResult
}

func NewMemorySessionRepository() *MemorySessionRepository {
	return &MemorySessionRepository{data: make(map[string]model.TranslationResult)}
}

func (r *MemorySessionRepository) Save(sessionID string, result model.TranslationResult) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[sessionID] = result
	return nil
}

func (r *MemorySessionRepository) Get(sessionID string) (model.TranslationResult, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	res, ok := r.data[sessionID]
	if !ok {
		return model.TranslationResult{}, fmt.Errorf("session %s not found", sessionID)
	}
	return res, nil
}
