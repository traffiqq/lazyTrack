package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (a *App) View() string {
	if !a.ready {
		return "Loading..."
	}

	// Render overlays if active
	if a.notifDialog.active {
		return a.notifDialog.View(a.width, a.height)
	}
	if a.finderDialog.active {
		return a.finderDialog.View(a.width, a.height)
	}
	if a.projectPicker.active {
		return a.projectPicker.View(a.width, a.height)
	}
	if a.showHelp {
		return renderHelp(a.width, a.height)
	}
	if a.issueDialog.active {
		return a.issueDialog.View(a.width, a.height)
	}

	var panels string
	panelHeight := a.height - 3
	if !a.listCollapsed {
		panelHeight-- // filter bar takes 1 line inside the list panel
	}
	hasComments := a.hasComments()

	// Determine detail panel title
	detailTitle := iconFile + " Detail"
	if a.selected != nil {
		detailTitle = iconFile + " " + a.selected.IDReadable
	}

	// Comments panel title with count
	commentsTitle := ""
	if hasComments {
		commentsTitle = fmt.Sprintf("%s Comments (%d)", iconComment, len(a.selected.Comments))
	}

	if a.listCollapsed {
		if a.commenting {
			innerWidth := a.width - 2
			commentContent := lipgloss.NewStyle().Padding(1, 2).Render(
				a.commentInput.View() + "\n\n" +
					hintDescStyle.Render("ctrl+s: submit  esc: cancel"),
			)
			panels = renderTitledPanel(iconFile+" Add Comment", commentContent, innerWidth, panelHeight, true, lipgloss.Color("99"))
		} else if hasComments {
			detailOuter := a.width / 2
			commentsOuter := a.width - detailOuter
			innerDetailWidth := detailOuter - 2
			innerCommentsWidth := commentsOuter - 2

			detailPanel := renderTitledPanel(detailTitle, a.detail.View(), innerDetailWidth, panelHeight, a.focus == detailPane, lipgloss.Color("69"))
			commentsPanel := renderTitledPanel(commentsTitle, a.comments.View(), innerCommentsWidth, panelHeight, a.focus == commentsPane, lipgloss.Color("99"))
			panels = lipgloss.JoinHorizontal(lipgloss.Top, detailPanel, commentsPanel)
		} else {
			innerWidth := a.width - 2
			panels = renderTitledPanel(detailTitle, a.detail.View(), innerWidth, panelHeight, true, lipgloss.Color("69"))
		}
	} else {
		listWidth := int(float64(a.width) * a.listRatio)
		innerListWidth := listWidth - 2

		filterBar := a.renderFilterBar(innerListWidth)
		listContent := filterBar + "\n" + a.list.View()
		leftPanel := renderTitledPanel(iconList+" Issues", listContent, innerListWidth, panelHeight, a.focus == listPane, lipgloss.Color("78"))

		if a.commenting {
			detailWidth := a.width - listWidth
			innerDetailWidth := detailWidth - 2
			commentContent := lipgloss.NewStyle().Padding(1, 2).Render(
				a.commentInput.View() + "\n\n" +
					hintDescStyle.Render("ctrl+s: submit  esc: cancel"),
			)
			rightPanel := renderTitledPanel(iconFile+" Add Comment", commentContent, innerDetailWidth, panelHeight, true, lipgloss.Color("99"))
			panels = lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel)
		} else if hasComments {
			remaining := a.width - listWidth
			detailOuter := remaining / 2
			commentsOuter := remaining - detailOuter
			innerDetailWidth := detailOuter - 2
			innerCommentsWidth := commentsOuter - 2

			detailPanel := renderTitledPanel(detailTitle, a.detail.View(), innerDetailWidth, panelHeight, a.focus == detailPane, lipgloss.Color("69"))
			commentsPanel := renderTitledPanel(commentsTitle, a.comments.View(), innerCommentsWidth, panelHeight, a.focus == commentsPane, lipgloss.Color("99"))
			panels = lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, detailPanel, commentsPanel)
		} else {
			detailWidth := a.width - listWidth
			innerDetailWidth := detailWidth - 2
			rightPanel := renderTitledPanel(detailTitle, a.detail.View(), innerDetailWidth, panelHeight, a.focus == detailPane, lipgloss.Color("69"))
			panels = lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel)
		}
	}

	var bottom string
	if a.searching {
		bottom = a.searchInput.View()
	} else if a.settingState {
		bottom = a.stateInput.View()
	} else if a.assigning {
		bottom = a.assignInput.View()
	} else if a.goingToIssue {
		bottom = a.gotoInput.View()
	} else if a.confirmDelete && a.selected != nil {
		bottom = errorStyle.Render(fmt.Sprintf("Delete %s? (y/n)", a.selected.IDReadable))
	} else {
		bottom = a.renderStatusBar()
	}

	layout := lipgloss.JoinVertical(lipgloss.Left, panels, bottom)

	if a.leaderActive {
		popup := a.renderLeaderPopup()
		layout = overlayBottomRight(layout, popup, a.width, a.height)
	}

	return layout
}

func (a *App) hasComments() bool {
	return a.selected != nil && len(a.selected.Comments) > 0
}

func (a *App) renderFilterBar(width int) string {
	type chip struct {
		key    string
		label  string
		active bool
	}

	chips := []chip{
		{"1", "Me", a.filterMe},
		{"2", "Bug", a.filterBug},
		{"3", "Task", a.filterTask},
	}

	var parts []string
	for _, c := range chips {
		mark := "☐"
		style := filterInactiveStyle
		if c.active {
			mark = "☑"
			style = filterActiveStyle
		}
		parts = append(parts, style.Render(fmt.Sprintf("%s:%s %s", c.key, mark, c.label)))
	}

	// Gap between Me (assignee) and Bug/Task (type) filters
	bar := parts[0] + "  " + parts[1] + " " + parts[2]

	return lipgloss.NewStyle().Width(width).Render(bar)
}

func (a *App) renderLeaderPopup() string {
	// Two-column layout, alphabetically sorted
	hints := leaderHints
	mid := (len(hints) + 1) / 2

	var rows []string
	for i := 0; i < mid; i++ {
		left := formatKeyHint(hints[i].key, hints[i].desc)
		leftPad := lipgloss.NewStyle().Width(14).Render(left)
		right := ""
		if i+mid < len(hints) {
			right = formatKeyHint(hints[i+mid].key, hints[i+mid].desc)
		}
		rows = append(rows, leftPad+right)
	}

	content := strings.Join(rows, "\n")

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("99")).
		Padding(0, 1).
		Render(content)
}

func overlayBottomRight(bg, fg string, width, height int) string {
	bgLines := strings.Split(bg, "\n")
	fgLines := strings.Split(fg, "\n")

	fgWidth := lipgloss.Width(fg)
	startRow := len(bgLines) - len(fgLines) - 1 // -1 to stay above status bar
	startCol := width - fgWidth

	if startRow < 0 {
		startRow = 0
	}
	if startCol < 0 {
		startCol = 0
	}

	for i, fgLine := range fgLines {
		row := startRow + i
		if row >= len(bgLines) {
			break
		}
		bgLine := bgLines[row]
		// Pad background line to width if needed
		bgLineWidth := lipgloss.Width(bgLine)
		if bgLineWidth < startCol {
			bgLine += strings.Repeat(" ", startCol-bgLineWidth)
		}
		// Replace segment: keep left part, overlay fg, keep right part
		leftPart := ansiTruncate(bgLine, startCol)
		bgLines[row] = leftPart + fgLine
	}

	return strings.Join(bgLines, "\n")
}

// ansiTruncate truncates a string to maxWidth visible characters,
// preserving ANSI escape sequences.
func ansiTruncate(s string, maxWidth int) string {
	var result strings.Builder
	visible := 0
	inEscape := false

	for _, r := range s {
		if r == '\x1b' {
			inEscape = true
			result.WriteRune(r)
			continue
		}
		if inEscape {
			result.WriteRune(r)
			if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') {
				inEscape = false
			}
			continue
		}
		if visible >= maxWidth {
			break
		}
		result.WriteRune(r)
		visible++
	}

	return result.String()
}

func (a *App) resizePanels() {
	panelHeight := a.height - 5
	hasComments := a.hasComments()

	if a.listCollapsed {
		if hasComments {
			detailOuter := a.width / 2
			commentsOuter := a.width - detailOuter
			a.detail.Width = detailOuter - 4
			a.detail.Height = panelHeight
			a.comments.Width = commentsOuter - 4
			a.comments.Height = panelHeight
		} else {
			a.detail.Width = a.width - 4
			a.detail.Height = panelHeight
		}
	} else {
		listOuter := int(float64(a.width) * a.listRatio)
		listWidth := listOuter - 4
		a.list.SetSize(listWidth, panelHeight-1) // -1 for filter bar line

		if hasComments {
			remaining := a.width - listOuter
			detailOuter := remaining / 2
			commentsOuter := remaining - detailOuter
			a.detail.Width = detailOuter - 4
			a.detail.Height = panelHeight
			a.comments.Width = commentsOuter - 4
			a.comments.Height = panelHeight
		} else {
			detailWidth := a.width - listOuter - 4
			a.detail.Width = detailWidth
			a.detail.Height = panelHeight
		}
	}
}
