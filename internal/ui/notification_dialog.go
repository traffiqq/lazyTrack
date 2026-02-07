package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/cf/lazytrack/internal/model"
)

type NotificationDialog struct {
	active        bool
	submitted     bool
	selectedIssue *model.Issue
	issues        []model.Issue
	cursor        int
	loading       bool
	err           string
	lastChecked   int64
}

func NewNotificationDialog() NotificationDialog {
	return NotificationDialog{}
}

func (d *NotificationDialog) Open(lastChecked int64) {
	d.active = true
	d.submitted = false
	d.selectedIssue = nil
	d.issues = nil
	d.cursor = 0
	d.loading = true
	d.err = ""
	d.lastChecked = lastChecked
}

func (d *NotificationDialog) Close() {
	d.active = false
}

func (d *NotificationDialog) SetResults(issues []model.Issue) {
	d.issues = issues
	d.cursor = 0
	d.loading = false
	d.err = ""
}

func (d *NotificationDialog) SetError(errStr string) {
	d.loading = false
	d.err = errStr
}

func (d NotificationDialog) Update(msg tea.Msg) (NotificationDialog, tea.Cmd) {
	if !d.active {
		return d, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			d.Close()
			return d, nil
		case "enter":
			if len(d.issues) > 0 {
				issue := d.issues[d.cursor]
				d.selectedIssue = &issue
				d.submitted = true
				d.Close()
			}
			return d, nil
		case "up", "k":
			if d.cursor > 0 {
				d.cursor--
			}
			return d, nil
		case "down", "j":
			if d.cursor < len(d.issues)-1 {
				d.cursor++
			}
			return d, nil
		}
	}

	return d, nil
}

func (d *NotificationDialog) View(width, height int) string {
	if !d.active {
		return ""
	}

	dialogWidth := width * 3 / 5
	if dialogWidth < 50 {
		dialogWidth = 50
	}
	dialogHeight := height * 7 / 10
	if dialogHeight < 15 {
		dialogHeight = 15
	}
	if dialogHeight > height-2 {
		dialogHeight = height - 2
	}

	contentWidth := dialogWidth - 6

	var b strings.Builder

	b.WriteString(titleStyle.Render("Mentions") + "\n\n")

	resultsHeight := dialogHeight - 7
	if resultsHeight < 3 {
		resultsHeight = 3
	}

	if d.err != "" {
		b.WriteString(errorStyle.Render("Error: "+d.err) + "\n")
	} else if d.loading {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("Loading mentions...") + "\n")
	} else if len(d.issues) == 0 {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("No mentions found") + "\n")
	} else {
		normalStyle := lipgloss.NewStyle().Width(contentWidth)
		selectedStyle := lipgloss.NewStyle().
			Width(contentWidth).
			Background(lipgloss.Color("237")).
			Foreground(lipgloss.Color("255"))
		newBadge := lipgloss.NewStyle().
			Foreground(lipgloss.Color("220")).
			Bold(true)

		start := 0
		if d.cursor >= resultsHeight {
			start = d.cursor - resultsHeight + 1
		}
		end := start + resultsHeight
		if end > len(d.issues) {
			end = len(d.issues)
		}

		for i := start; i < end; i++ {
			issue := d.issues[i]
			badge := "     "
			if issue.Updated > d.lastChecked {
				badge = newBadge.Render("[NEW]")
			}
			line := fmt.Sprintf("%s %-12s %s", badge, issue.IDReadable, issue.Summary)
			if lipgloss.Width(line) > contentWidth {
				maxSummary := contentWidth - lipgloss.Width(badge) - 14
				if maxSummary < 0 {
					maxSummary = 0
				}
				summary := issue.Summary
				if len(summary) > maxSummary && maxSummary > 1 {
					summary = summary[:maxSummary-1] + "…"
				}
				line = fmt.Sprintf("%s %-12s %s", badge, issue.IDReadable, summary)
			}
			if i == d.cursor {
				b.WriteString(selectedStyle.Render(line) + "\n")
			} else {
				b.WriteString(normalStyle.Render(line) + "\n")
			}
		}
	}

	b.WriteString("\n")
	hint := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	b.WriteString(hint.Render("j/k: navigate  enter: open  esc: close"))

	dialogStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("99")).
		Padding(1, 2).
		Width(dialogWidth).
		Height(dialogHeight)

	dialog := dialogStyle.Render(b.String())

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, dialog)
}
