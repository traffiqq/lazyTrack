package ui

import (
	"strings"
	"testing"
)

func TestFormatKeyHint(t *testing.T) {
	result := formatKeyHint("j/k", "navigate")
	if !strings.Contains(result, "j/k") {
		t.Error("expected key text in output")
	}
	if !strings.Contains(result, "navigate") {
		t.Error("expected description text in output")
	}
}

func TestFormatHints(t *testing.T) {
	hints := []keyHint{
		{"a", "first"},
		{"b", "second"},
	}
	result := formatHints(hints)
	if !strings.Contains(result, "a") {
		t.Error("expected first key")
	}
	if !strings.Contains(result, "second") {
		t.Error("expected second description")
	}
}

func TestModeHints(t *testing.T) {
	tests := []struct {
		name       string
		commenting bool
		focus      pane
		wantKey    string
		wantAbsent string
	}{
		{"list mode", false, listPane, "navigate", "scroll"},
		{"detail mode", false, detailPane, "scroll", "navigate"},
		{"commenting mode", true, detailPane, "submit", "navigate"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hints := modeHints(tt.commenting, tt.focus)
			joined := formatHints(hints)
			if !strings.Contains(joined, tt.wantKey) {
				t.Errorf("expected %q in hints, got %q", tt.wantKey, joined)
			}
			if strings.Contains(joined, tt.wantAbsent) {
				t.Errorf("did not expect %q in hints, got %q", tt.wantAbsent, joined)
			}
		})
	}
}
