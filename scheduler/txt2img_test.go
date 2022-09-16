package scheduler

import (
	"testing"
)

func TestClean(t *testing.T) {
	tests := []struct {
		prompt   string
		expected string
	}{
		{"", ""},
		{"valid test", "valid test"},
		{"valid test   with    extra    spaces", "valid test with extra spaces"},
	}

	for _, tt := range tests {
		msg := clean(tt.prompt)
		if msg != tt.expected {
			t.Fatalf(`Got = %q, want %q, error`, msg, tt.expected)
		}
	}
}

func TestRemoveMentions(t *testing.T) {
	tests := []struct {
		prompt   string
		expected string
	}{
		{"", ""},
		{"valid test", "valid test"},
		{"valid test with @mentions in the middle", "valid test with  in the middle"},
	}

	for _, tt := range tests {
		msg := removeMentions(tt.prompt)
		if msg != tt.expected {
			t.Fatalf(`Got = %q, want %q, error`, msg, tt.expected)
		}
	}
}

func TestRemoveCommands(t *testing.T) {
	tests := []struct {
		prompt   string
		expected string
	}{
		{"", ""},
		{"valid test", "valid test"},
		{"valid test with /commands in the middle", "valid test with  in the middle"},
	}

	for _, tt := range tests {
		msg := removeCommands(tt.prompt)
		if msg != tt.expected {
			t.Fatalf(`Got = %q, want %q, error`, msg, tt.expected)
		}
	}
}
