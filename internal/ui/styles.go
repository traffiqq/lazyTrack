package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	panelStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240"))

	focusedPanelStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("69"))

	statusBarStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Padding(0, 1)

	titleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("99"))

	errorStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("196")).
		Padding(0, 1)
)

// Nerd Font icon constants.
const (
	iconList = "\uF0CA" // nf-fa-list_ul
	iconFile = "\uF15C" // nf-fa-file_text_o
	iconApp  = "\uF188" // nf-fa-bug
)

// Chrome styles for the status bar key hints.
var (
	keyStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("220")) // yellow

	hintDescStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("245")) // gray
)

// keyHint pairs a key with its description for the status bar.
type keyHint struct {
	key  string
	desc string
}

// Mode-specific key hint sets, ordered by importance (rightmost dropped first on overflow).
var (
	listHints = []keyHint{
		{"j/k", "navigate"},
		{"enter", "open"},
		{"/", "search"},
		{"f", "find"},
		{"c", "create"},
		{"p", "project"},
		{"#", "goto"},
		{"ctrl+e", "collapse"},
		{"H/L", "resize"},
		{"?", "help"},
		{"q", "quit"},
	}
	detailHints = []keyHint{
		{"j/k", "scroll"},
		{"e", "edit"},
		{"d", "delete"},
		{"C", "comment"},
		{"s", "state"},
		{"a", "assign"},
		{"p", "project"},
		{"#", "goto"},
		{"tab", "switch"},
		{"ctrl+e", "collapse"},
		{"H/L", "resize"},
		{"q", "quit"},
	}
	commentingHints = []keyHint{
		{"ctrl+d", "submit"},
		{"esc", "cancel"},
	}
)

// formatKeyHint renders a single key hint with colored key and gray description.
func formatKeyHint(key, desc string) string {
	return keyStyle.Render(key) + hintDescStyle.Render(":"+desc)
}

// formatHints joins a slice of key hints into a spaced string.
func formatHints(hints []keyHint) string {
	parts := make([]string, len(hints))
	for i, h := range hints {
		parts[i] = formatKeyHint(h.key, h.desc)
	}
	return strings.Join(parts, "  ")
}

// modeHints returns the appropriate key hints for the current UI mode.
func modeHints(commenting bool, focus pane) []keyHint {
	switch {
	case commenting:
		return commentingHints
	case focus == listPane:
		return listHints
	default:
		return detailHints
	}
}

// stateColor returns a colored string for common YouTrack issue states.
func stateColor(state string) string {
	switch state {
	case "Open", "Submitted", "Reopened":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Render(state) // yellow
	case "In Progress", "In Review":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("69")).Render(state) // blue
	case "Fixed", "Done", "Verified", "Resolved":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("78")).Render(state) // green
	case "Won't Fix", "Duplicate", "Obsolete", "Incomplete":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render(state) // gray
	default:
		return state
	}
}
