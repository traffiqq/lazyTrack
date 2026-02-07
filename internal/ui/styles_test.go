package ui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
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

func TestRenderTitledPanel(t *testing.T) {
	result := renderTitledPanel(iconList+" Issues", "hello world", 30, 5, true, lipgloss.Color("78"))
	lines := strings.Split(result, "\n")

	// First line: custom top border containing the title
	if !strings.Contains(lines[0], "Issues") {
		t.Errorf("first line should contain title text, got: %q", lines[0])
	}
	if !strings.Contains(lines[0], "╭") {
		t.Errorf("first line should contain top-left corner, got: %q", lines[0])
	}
	if !strings.Contains(lines[0], "╮") {
		t.Errorf("first line should contain top-right corner, got: %q", lines[0])
	}

	// Content should appear in the body
	if !strings.Contains(result, "hello world") {
		t.Error("expected content in output")
	}

	// Last line should have bottom border
	lastLine := lines[len(lines)-1]
	if !strings.Contains(lastLine, "╰") {
		t.Errorf("last line should contain bottom-left corner, got: %q", lastLine)
	}
}

func TestRenderTitledPanelUnfocused(t *testing.T) {
	lipgloss.SetColorProfile(termenv.ANSI256)
	defer lipgloss.SetColorProfile(termenv.Ascii)
	focused := renderTitledPanel("Test", "x", 20, 3, true, lipgloss.Color("78"))
	unfocused := renderTitledPanel("Test", "x", 20, 3, false, lipgloss.Color("78"))

	if !strings.Contains(focused, "Test") {
		t.Error("focused panel should contain title")
	}
	if !strings.Contains(unfocused, "Test") {
		t.Error("unfocused panel should contain title")
	}
	if focused == unfocused {
		t.Error("focused and unfocused panels should render differently")
	}
}

func TestRenderTitledPanelWidthConsistency(t *testing.T) {
	result := renderTitledPanel("Test Title", "content", 40, 5, true, lipgloss.Color("78"))
	lines := strings.Split(result, "\n")
	if len(lines) < 2 {
		t.Fatal("expected at least 2 lines")
	}
	topWidth := lipgloss.Width(lines[0])
	bodyWidth := lipgloss.Width(lines[1])
	if topWidth != bodyWidth {
		t.Errorf("top border width %d != body width %d", topWidth, bodyWidth)
	}
}
