package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/cf/lazytrack/internal/model"
)

// ProjectPickerDialog is a centered popup for selecting a project.
type ProjectPickerDialog struct {
	projects        []model.Project
	cursor          int
	active          bool
	submitted       bool
	selectedProject *model.Project
}

func NewProjectPickerDialog() ProjectPickerDialog {
	return ProjectPickerDialog{}
}

// Open activates the dialog with the given projects.
// Entry 0 in the display is "(All Projects)"; projects start at index 1.
func (d *ProjectPickerDialog) Open(projects []model.Project) {
	d.projects = projects
	d.cursor = 0
	d.active = true
	d.submitted = false
	d.selectedProject = nil
}

func (d *ProjectPickerDialog) Close() {
	d.active = false
}

// totalItems returns the number of selectable entries (All Projects + len(projects)).
func (d *ProjectPickerDialog) totalItems() int {
	return 1 + len(d.projects)
}

func (d *ProjectPickerDialog) Update(msg tea.Msg) (ProjectPickerDialog, tea.Cmd) {
	if !d.active {
		return *d, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			d.Close()
			return *d, nil
		case "enter":
			d.submitted = true
			if d.cursor == 0 {
				d.selectedProject = nil
			} else {
				proj := d.projects[d.cursor-1]
				d.selectedProject = &proj
			}
			d.Close()
			return *d, nil
		case "up", "k":
			if d.cursor > 0 {
				d.cursor--
			}
			return *d, nil
		case "down", "j":
			if d.cursor < d.totalItems()-1 {
				d.cursor++
			}
			return *d, nil
		}
	}

	return *d, nil
}

func (d *ProjectPickerDialog) View(width, height int) string {
	if !d.active {
		return ""
	}

	dialogWidth := width * 2 / 5
	if dialogWidth < 40 {
		dialogWidth = 40
	}

	contentWidth := dialogWidth - 6

	var b strings.Builder

	b.WriteString(titleStyle.Render("Select Project") + "\n\n")

	normalStyle := lipgloss.NewStyle().Width(contentWidth)
	selectedStyle := lipgloss.NewStyle().
		Width(contentWidth).
		Background(lipgloss.Color("237")).
		Foreground(lipgloss.Color("255"))

	// (All Projects) entry
	line := "(All Projects)"
	if d.cursor == 0 {
		b.WriteString(selectedStyle.Render(line) + "\n")
	} else {
		b.WriteString(normalStyle.Render(line) + "\n")
	}

	// Project entries
	for i, proj := range d.projects {
		line := fmt.Sprintf("%s (%s)", proj.Name, proj.ShortName)
		if len(line) > contentWidth {
			line = line[:contentWidth-1] + "…"
		}
		if d.cursor == i+1 {
			b.WriteString(selectedStyle.Render(line) + "\n")
		} else {
			b.WriteString(normalStyle.Render(line) + "\n")
		}
	}

	b.WriteString("\n")
	hint := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	b.WriteString(hint.Render("j/k: navigate  enter: select  esc: cancel"))

	dialogStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("99")).
		Padding(1, 2).
		Width(dialogWidth)

	dialog := dialogStyle.Render(b.String())

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, dialog)
}
