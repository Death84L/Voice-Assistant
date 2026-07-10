package iface

import "VA/internal/model"

// SessionRepository persists the result of each processed utterance.
type SessionRepository interface {
	Save(sessionID string, result model.TranslationResult) error
	Get(sessionID string) (model.TranslationResult, error)
}
