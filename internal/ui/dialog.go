package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/cf/lazytrack/internal/model"
)

// CreateDialog is a modal form for creating a new issue.
type CreateDialog struct {
	projects     []model.Project
	projectIndex int
	summaryInput textinput.Model
	descInput    textarea.Model
	focusIndex   int // 0=project, 1=summary, 2=description
	active       bool
	submitted    bool
}

func NewCreateDialog() CreateDialog {
	si := textinput.New()
	si.Placeholder = "Issue summary"
	si.Prompt = ""
	si.CharLimit = 200

	di := textarea.New()
	di.Placeholder = "Description (optional)"
	di.SetHeight(5)
	di.CharLimit = 5000

	return CreateDialog{
		summaryInput: si,
		descInput:    di,
	}
}

func (d *CreateDialog) SetProjects(projects []model.Project) {
	d.projects = projects
	d.projectIndex = 0
}

func (d *CreateDialog) Open() tea.Cmd {
	d.active = true
	d.submitted = false
	d.focusIndex = 1
	d.summaryInput.SetValue("")
	d.descInput.SetValue("")
	return d.summaryInput.Focus()
}

func (d *CreateDialog) Close() {
	d.active = false
	d.summaryInput.Blur()
	d.descInput.Blur()
}

func (d *CreateDialog) Values() (projectID, summary, description string) {
	if len(d.projects) > 0 {
		projectID = d.projects[d.projectIndex].ID
	}
	return projectID, d.summaryInput.Value(), d.descInput.Value()
}

func (d *CreateDialog) Update(msg tea.Msg) (CreateDialog, tea.Cmd) {
	if !d.active {
		return *d, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			d.Close()
			return *d, nil
		case "tab":
			d.focusIndex = (d.focusIndex + 1) % 3
			return *d, d.updateFocus()
		case "shift+tab":
			d.focusIndex = (d.focusIndex + 2) % 3
			return *d, d.updateFocus()
		case "left":
			if d.focusIndex == 0 && len(d.projects) > 0 {
				d.projectIndex = (d.projectIndex + len(d.projects) - 1) % len(d.projects)
				return *d, nil
			}
		case "right":
			if d.focusIndex == 0 && len(d.projects) > 0 {
				d.projectIndex = (d.projectIndex + 1) % len(d.projects)
				return *d, nil
			}
		case "ctrl+d":
			if d.summaryInput.Value() != "" {
				d.submitted = true
				d.Close()
			}
			return *d, nil
		}
	}

	var cmd tea.Cmd
	switch d.focusIndex {
	case 1:
		d.summaryInput, cmd = d.summaryInput.Update(msg)
	case 2:
		d.descInput, cmd = d.descInput.Update(msg)
	}
	return *d, cmd
}

func (d *CreateDialog) updateFocus() tea.Cmd {
	d.summaryInput.Blur()
	d.descInput.Blur()
	switch d.focusIndex {
	case 1:
		return d.summaryInput.Focus()
	case 2:
		return d.descInput.Focus()
	}
	return nil
}

func (d *CreateDialog) View(width, height int) string {
	if !d.active {
		return ""
	}

	dialogWidth := width * 3 / 5
	if dialogWidth < 40 {
		dialogWidth = 40
	}

	var b strings.Builder

	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("99"))
	b.WriteString(title.Render("Create Issue") + "\n\n")

	// Project selector
	label := lipgloss.NewStyle().Bold(d.focusIndex == 0).Foreground(lipgloss.Color("69"))
	projName := "(no projects)"
	if len(d.projects) > 0 {
		p := d.projects[d.projectIndex]
		projName = fmt.Sprintf("< %s (%s) >", p.Name, p.ShortName)
	}
	b.WriteString(label.Render("Project: ") + projName + "\n\n")

	// Summary
	summaryLabel := lipgloss.NewStyle().Bold(d.focusIndex == 1).Foreground(lipgloss.Color("69"))
	b.WriteString(summaryLabel.Render("Summary:") + "\n")
	b.WriteString(d.summaryInput.View() + "\n\n")

	// Description
	descLabel := lipgloss.NewStyle().Bold(d.focusIndex == 2).Foreground(lipgloss.Color("69"))
	b.WriteString(descLabel.Render("Description:") + "\n")
	b.WriteString(d.descInput.View() + "\n\n")

	hint := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	b.WriteString(hint.Render("tab: next field  left/right: project  ctrl+d: submit  esc: cancel"))

	content := b.String()

	dialogStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("99")).
		Padding(1, 2).
		Width(dialogWidth)

	dialog := dialogStyle.Render(content)

	// Center the dialog
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, dialog)
}

// EditDialog is a modal form for editing an existing issue.
type EditDialog struct {
	summaryInput textinput.Model
	descInput    textarea.Model
	focusIndex   int // 0=summary, 1=description
	active       bool
	submitted    bool
	issueID      string
}

func NewEditDialog() EditDialog {
	si := textinput.New()
	si.Prompt = ""
	si.CharLimit = 200

	di := textarea.New()
	di.SetHeight(5)
	di.CharLimit = 5000

	return EditDialog{
		summaryInput: si,
		descInput:    di,
	}
}

func (d *EditDialog) Open(issueID, summary, description string) tea.Cmd {
	d.active = true
	d.submitted = false
	d.issueID = issueID
	d.focusIndex = 0
	d.summaryInput.SetValue(summary)
	d.descInput.SetValue(description)
	return d.summaryInput.Focus()
}

func (d *EditDialog) Close() {
	d.active = false
	d.summaryInput.Blur()
	d.descInput.Blur()
}

func (d *EditDialog) Values() (summary, description string) {
	return d.summaryInput.Value(), d.descInput.Value()
}

func (d *EditDialog) Update(msg tea.Msg) (EditDialog, tea.Cmd) {
	if !d.active {
		return *d, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			d.Close()
			return *d, nil
		case "tab":
			d.focusIndex = (d.focusIndex + 1) % 2
			return *d, d.updateFocus()
		case "shift+tab":
			d.focusIndex = (d.focusIndex + 1) % 2
			return *d, d.updateFocus()
		case "ctrl+d":
			if d.summaryInput.Value() != "" {
				d.submitted = true
				d.Close()
			}
			return *d, nil
		}
	}

	var cmd tea.Cmd
	switch d.focusIndex {
	case 0:
		d.summaryInput, cmd = d.summaryInput.Update(msg)
	case 1:
		d.descInput, cmd = d.descInput.Update(msg)
	}
	return *d, cmd
}

func (d *EditDialog) updateFocus() tea.Cmd {
	d.summaryInput.Blur()
	d.descInput.Blur()
	switch d.focusIndex {
	case 0:
		return d.summaryInput.Focus()
	case 1:
		return d.descInput.Focus()
	}
	return nil
}

func (d *EditDialog) View(width, height int) string {
	if !d.active {
		return ""
	}

	dialogWidth := width * 3 / 5
	if dialogWidth < 40 {
		dialogWidth = 40
	}

	var b strings.Builder

	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("99"))
	b.WriteString(title.Render("Edit Issue — "+d.issueID) + "\n\n")

	summaryLabel := lipgloss.NewStyle().Bold(d.focusIndex == 0).Foreground(lipgloss.Color("69"))
	b.WriteString(summaryLabel.Render("Summary:") + "\n")
	b.WriteString(d.summaryInput.View() + "\n\n")

	descLabel := lipgloss.NewStyle().Bold(d.focusIndex == 1).Foreground(lipgloss.Color("69"))
	b.WriteString(descLabel.Render("Description:") + "\n")
	b.WriteString(d.descInput.View() + "\n\n")

	hint := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	b.WriteString(hint.Render("tab: next field  ctrl+d: submit  esc: cancel"))

	content := b.String()

	dialogStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("99")).
		Padding(1, 2).
		Width(dialogWidth)

	dialog := dialogStyle.Render(content)

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, dialog)
}
