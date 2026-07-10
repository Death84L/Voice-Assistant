package service

import (
	"regexp"
	"strings"
)

// TextCleaner has one job: strip filler words before translation
// so "uh", "um", "you know" don't get translated literally.
type TextCleaner struct {
	fillerPattern *regexp.Regexp
}

func NewTextCleaner() *TextCleaner {
	fillers := []string{`\buh\b`, `\bumm?\b`, `\byou know\b`, `\blike\b`}
	pattern := regexp.MustCompile(`(?i)(` + strings.Join(fillers, "|") + `)`)
	return &TextCleaner{fillerPattern: pattern}
}

func (c *TextCleaner) RemoveFillers(text string) string {
	cleaned := c.fillerPattern.ReplaceAllString(text, "")
	cleaned = regexp.MustCompile(`\s+`).ReplaceAllString(cleaned, " ")
	return strings.TrimSpace(cleaned)
}
