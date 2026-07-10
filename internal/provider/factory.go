package provider

import "VA/internal/iface"

// ProviderFactory creates concrete provider implementations.
type ProviderFactory struct {
	groqKey string
}

func NewProviderFactory(groqKey string) *ProviderFactory {
	return &ProviderFactory{
		groqKey: groqKey,
	}
}

func (f *ProviderFactory) NewTranscriber(name string) iface.Transcriber {
	switch name {
	case "grok":
		return NewGrokASR(f.groqKey)
	default:
		return NewGrokASR(f.groqKey)
	}
}

func (f *ProviderFactory) NewTranslator(name string) iface.Translator {
	switch name {
	case "grok":
		return NewGrokTranslator(f.groqKey)
	default:
		return NewGrokTranslator(f.groqKey)
	}
}

func (f *ProviderFactory) NewSynthesizer(name string) iface.Synthesizer {
	switch name {
	case "grok":
		return NewGrokTTS(f.groqKey)
	default:
		return NewGrokTTS(f.groqKey)
	}
}