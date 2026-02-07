package ui

import (
	"fmt"

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

	// Determine detail panel title
	detailTitle := iconFile + " Detail"
	if a.selected != nil {
		detailTitle = iconFile + " " + a.selected.IDReadable
	}

	if a.listCollapsed {
		innerWidth := a.width - 2
		if a.commenting {
			commentContent := lipgloss.NewStyle().Padding(1, 2).Render(
				a.commentInput.View() + "\n\n" +
					hintDescStyle.Render("ctrl+s: submit  esc: cancel"),
			)
			panels = renderTitledPanel(iconFile+" Add Comment", commentContent, innerWidth, panelHeight, true, lipgloss.Color("99"))
		} else {
			panels = renderTitledPanel(detailTitle, a.detail.View(), innerWidth, panelHeight, true, lipgloss.Color("69"))
		}
	} else {
		listWidth := int(float64(a.width) * a.listRatio)
		detailWidth := a.width - listWidth
		innerListWidth := listWidth - 2
		innerDetailWidth := detailWidth - 2

		leftPanel := renderTitledPanel(iconList+" Issues", a.list.View(), innerListWidth, panelHeight, a.focus == listPane, lipgloss.Color("78"))

		if a.commenting {
			commentContent := lipgloss.NewStyle().Padding(1, 2).Render(
				a.commentInput.View() + "\n\n" +
					hintDescStyle.Render("ctrl+s: submit  esc: cancel"),
			)
			rightPanel := renderTitledPanel(iconFile+" Add Comment", commentContent, innerDetailWidth, panelHeight, true, lipgloss.Color("99"))
			panels = lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel)
		} else {
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

	return lipgloss.JoinVertical(lipgloss.Left, panels, bottom)
}

func (a *App) resizePanels() {
	panelHeight := a.height - 5

	if a.listCollapsed {
		a.detail.Width = a.width - 4
		a.detail.Height = panelHeight
	} else {
		listOuter := int(float64(a.width) * a.listRatio)
		listWidth := listOuter - 4
		detailWidth := a.width - listOuter - 4
		a.list.SetSize(listWidth, panelHeight)
		a.detail.Width = detailWidth
		a.detail.Height = panelHeight
	}
}
