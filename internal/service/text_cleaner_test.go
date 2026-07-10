package service_test

import (
	"testing"

	"VA/internal/service"
)

func TestTextCleaner_RemoveFillers(t *testing.T) {
	cleaner := service.NewTextCleaner()

	cases := []struct {
		input    string
		expected string
	}{
		{"uh can we move the meeting", "can we move the meeting"},
		{"you know I think we should reschedule", "I think we should reschedule"},
		{"no fillers here", "no fillers here"},
		{"um so basically like it works", "so basically it works"},
	}

	for _, c := range cases {
		got := cleaner.RemoveFillers(c.input)
		if got != c.expected {
			t.Errorf("input=%q expected=%q got=%q", c.input, c.expected, got)
		}
	}
}
