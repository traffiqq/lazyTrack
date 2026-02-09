package ui

import (
	"strings"
	"testing"
)

func TestRenderMarkdown_HeadingAndBold(t *testing.T) {
	input := "# Hello\n\nSome **bold** text here."
	out := renderMarkdown(input, 80)
	if out == "" {
		t.Fatal("expected non-empty output")
	}
	if !strings.Contains(out, "Hello") {
		t.Errorf("output should contain 'Hello', got: %q", out)
	}
	if !strings.Contains(out, "bold") {
		t.Errorf("output should contain 'bold', got: %q", out)
	}
}

func TestRenderMarkdown_List(t *testing.T) {
	input := "- item one\n- item two\n"
	out := renderMarkdown(input, 80)
	if !strings.Contains(out, "item one") {
		t.Errorf("output should contain 'item one', got: %q", out)
	}
	if !strings.Contains(out, "item two") {
		t.Errorf("output should contain 'item two', got: %q", out)
	}
}

func TestRenderMarkdown_Empty(t *testing.T) {
	out := renderMarkdown("", 80)
	if out != "" {
		t.Errorf("expected empty string for empty input, got: %q", out)
	}
}

func TestRenderMarkdown_WidthRespected(t *testing.T) {
	// Create a long sentence that should wrap at the given width.
	input := "This is a rather long sentence that should definitely be wrapped when the rendering width is set to a small value like thirty characters."
	width := 30
	out := renderMarkdown(input, width)
	for _, line := range strings.Split(out, "\n") {
		// Allow a small tolerance for ANSI escape sequences which add
		// non-visible characters to the line length.
		if len([]rune(line)) > width+20 {
			t.Errorf("line exceeds expected width %d (got %d runes): %q", width, len([]rune(line)), line)
		}
	}
}

func TestRenderMarkdown_PlainTextPassthrough(t *testing.T) {
	input := "Just some plain text with no markdown."
	out := renderMarkdown(input, 80)
	if !strings.Contains(out, "Just some plain text with no markdown.") {
		t.Errorf("plain text should pass through, got: %q", out)
	}
}
