package ui

import "github.com/charmbracelet/lipgloss"

func renderHelp(width, height int) string {
	helpText := `lazytrack - YouTrack TUI

Navigation:
  j/k, up/down   Navigate issues
  tab             Switch panels
  enter           Load issue detail

Actions:
  c           Create issue
  e           Edit issue
  d           Delete issue
  C           Add comment
  /           Search/filter
  f           Find issue
  s           Set state
  a           Assign issue

General:
  H/L, ctrl+←/→  Resize panels
  ctrl+e          Toggle issue list
  ?               Toggle help
  q               Quit`

	dialogStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("99")).
		Padding(1, 3)

	dialog := dialogStyle.Render(helpText)

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, dialog)
}
