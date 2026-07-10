package iface

// Translator converts English text into Arabic text.
type Translator interface {
	Translate(text string) (string, error)
}
