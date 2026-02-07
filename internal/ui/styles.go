package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	statusBarStyle = lipgloss.NewStyle().
		Background(lipgloss.Color("236")).
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
		{"r", "refresh"},
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
		{"r", "refresh"},
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

// renderTitledPanel renders content inside a rounded border with a styled title
// embedded in the top border line. Uses manual border construction because
// Lip Gloss v1.1.0 has no border-title API.
//
// Parameters:
//   - title: text to embed in the top border (e.g., "\uF0CA Issues")
//   - content: the panel body content
//   - innerWidth: content width (border adds 2 more for total outer width)
//   - height: content height passed to lipgloss.Height()
//   - focused: if true, border is blue (#69) and title is bold+titleColor;
//     if false, border is gray (#240) and title is dimmed (#245)
//   - titleColor: foreground color for the title when focused
func renderTitledPanel(title string, content string, innerWidth int, height int, focused bool, titleColor lipgloss.Color) string {
	borderColor := lipgloss.Color("240")
	if focused {
		borderColor = lipgloss.Color("69")
	}

	// Style the title text
	var styledTitle string
	if focused {
		styledTitle = lipgloss.NewStyle().Bold(true).Foreground(titleColor).Render(title)
	} else {
		styledTitle = lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render(title)
	}

	// Build custom top border: ╭─ Title ─────────╮
	// Outer width = innerWidth + 2 (left + right border characters)
	// Layout: ╭(1) ─(1) space(1) + styledTitle + space(1) + fill×─ + ╮(1)
	// So: 5 fixed chars + titleVisualWidth + fill = outerWidth
	outerWidth := innerWidth + 2
	bc := lipgloss.NewStyle().Foreground(borderColor)
	titleVisualWidth := lipgloss.Width(styledTitle)
	fill := outerWidth - 5 - titleVisualWidth
	if fill < 0 {
		fill = 0
	}
	topLine := bc.Render("╭─ ") + styledTitle + bc.Render(" "+strings.Repeat("─", fill)+"╮")

	// Render content with 3-sided border (top suppressed, replaced by custom line above).
	// BorderTop(false) suppresses the top border row. The custom topLine replaces it,
	// so the total rendered height remains height + 2 (same as a full 4-sided border).
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderTop(false).
		BorderForeground(borderColor).
		Width(innerWidth).
		Height(height)

	body := style.Render(content)

	return topLine + "\n" + body
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
