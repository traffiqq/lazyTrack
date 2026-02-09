package ui

import (
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/glamour/ansi"
)

func stringPtr(s string) *string { return &s }
func boolPtr(b bool) *bool       { return &b }
func uintPtr(u uint) *uint       { return &u }

// buildMarkdownStyle returns a custom glamour style config suited for the
// lazyTrack TUI panels.
func buildMarkdownStyle() ansi.StyleConfig {
	return ansi.StyleConfig{
		Document: ansi.StyleBlock{
			Margin: uintPtr(0),
		},
		Heading: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Bold:  boolPtr(true),
				Color: stringPtr("99"),
			},
		},
		H1: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Bold:  boolPtr(true),
				Color: stringPtr("99"),
			},
		},
		H2: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Bold:  boolPtr(true),
				Color: stringPtr("99"),
			},
		},
		H3: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Bold:  boolPtr(true),
				Color: stringPtr("99"),
			},
		},
		Strong: ansi.StylePrimitive{
			Bold: boolPtr(true),
		},
		Emph: ansi.StylePrimitive{
			Italic: boolPtr(true),
		},
		Code: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				BackgroundColor: stringPtr("236"),
			},
		},
		CodeBlock: ansi.StyleCodeBlock{
			StyleBlock: ansi.StyleBlock{
				StylePrimitive: ansi.StylePrimitive{
					BackgroundColor: stringPtr("236"),
				},
			},
		},
		Link: ansi.StylePrimitive{
			Color:     stringPtr("69"),
			Underline: boolPtr(true),
		},
		LinkText: ansi.StylePrimitive{
			Color:     stringPtr("69"),
			Underline: boolPtr(true),
		},
		Item: ansi.StylePrimitive{
			BlockPrefix: "\u2022 ",
		},
		BlockQuote: ansi.StyleBlock{
			Indent: uintPtr(1),
			StylePrimitive: ansi.StylePrimitive{
				Color: stringPtr("245"),
			},
		},
		HorizontalRule: ansi.StylePrimitive{
			Color:  stringPtr("240"),
			Format: "--------",
		},
		Paragraph: ansi.StyleBlock{},
	}
}

// renderMarkdown renders the given markdown text to a styled string suitable
// for terminal display at the specified width. Returns the plain text on
// rendering error and an empty string for empty input.
func renderMarkdown(text string, width int) string {
	if text == "" {
		return ""
	}

	style := buildMarkdownStyle()
	r, err := glamour.NewTermRenderer(
		glamour.WithStyles(style),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return text
	}

	out, err := r.Render(text)
	if err != nil {
		return text
	}

	return strings.TrimRight(out, " \t\n")
}
