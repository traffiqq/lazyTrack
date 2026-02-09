package ui

import "github.com/charmbracelet/lipgloss"

func renderHelp(width, height int) string {
	helpText := `lazytrack - YouTrack TUI

Navigation:
  j/k, up/down   Navigate issues
  tab             Cycle panels
  enter           Load issue detail

Direct Actions:
  /           Search/filter
  1/2/3       Filter: Me/Bug/Task
  #           Go to issue by number
  r           Refresh
  H/L         Resize panels

Leader Actions (space + key):
  space c     Create issue
  space e     Edit issue
  space d     Delete issue
  space m     Add comment
  space s     Set state
  space a     Assign issue
  space p     Select project
  space f     Find issue
  space n     Mentions
  space t     Toggle issue list

Dialogs & Comments:
  tab/shift+tab   Navigate fields
  ctrl+s          Submit
  esc             Cancel

General:
  ?               Toggle help
  q               Quit`

	dialogStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("99")).
		Padding(1, 3)

	dialog := dialogStyle.Render(helpText)

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, dialog)
}
