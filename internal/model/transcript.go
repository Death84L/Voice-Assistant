package model

// Transcript is the output of the ASR stage.
type Transcript struct {
	Text       string
	Confidence float64
	Language   string
}
