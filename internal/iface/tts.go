package iface

// Synthesizer converts Arabic text into speech audio bytes.
type Synthesizer interface {
	Synthesize(text string) ([]byte, error)
}
