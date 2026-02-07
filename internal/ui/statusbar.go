package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (a *App) renderStatusBar() string {
	if a.err != "" {
		return lipgloss.NewStyle().
			Background(lipgloss.Color("236")).
			Padding(0, 1).
			Width(a.width - 2).
			Render(errorStyle.Render("Error: " + a.err))
	}

	// Left side: app name + context
	left := titleStyle.Render(iconApp + " lazytrack")
	if a.activeProject != nil {
		left += hintDescStyle.Render(" | project: " + a.activeProject.ShortName)
	}
	if a.query != "" {
		left += hintDescStyle.Render(" | query: " + a.query)
	}
	if a.unreadMentionCount > 0 {
		left += mentionBadgeStyle.Render(fmt.Sprintf(" · %d mentions", a.unreadMentionCount))
	}
	if a.loading {
		left += keyStyle.Render(" | loading...")
	}

	// Right side: mode-aware hints
	hints := modeHints(a.commenting, a.focus)
	rightParts := make([]string, len(hints))
	for i, h := range hints {
		rightParts[i] = formatKeyHint(h.key, h.desc)
	}

	// Overflow: drop hints from right until content fits within available width
	leftWidth := lipgloss.Width(left)
	availWidth := a.width - 2 // content width inside statusBarStyle padding (0,1)
	for len(rightParts) > 0 {
		right := strings.Join(rightParts, "  ")
		if leftWidth+2+lipgloss.Width(right) <= availWidth {
			break
		}
		rightParts = rightParts[:len(rightParts)-1]
	}

	right := strings.Join(rightParts, "  ")
	rightWidth := lipgloss.Width(right)
	gap := availWidth - leftWidth - rightWidth
	if gap < 0 {
		gap = 0
	}

	content := left + strings.Repeat(" ", gap) + right
	return statusBarStyle.MaxWidth(a.width).Render(content)
}
