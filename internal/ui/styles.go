package ui

import "github.com/charmbracelet/lipgloss"

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
