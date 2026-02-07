package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/cf/lazytrack/internal/model"
)

type dialogMode int

const (
	modeCreate dialogMode = iota
	modeEdit
)

// issueField identifies which form field currently has focus.
type issueField int

const (
	fieldProject  issueField = iota // create only
	fieldType
	fieldState
	fieldAssignee
	fieldSummary
	fieldDescription
	fieldComments // edit only, when comments exist
)

// IssueDialog is a fullscreen form for creating or editing issues.
type IssueDialog struct {
	mode      dialogMode
	active    bool
	submitted bool

	// Issue being edited (nil for create)
	issueID string

	// Project selector (create mode only)
	projects       []model.Project
	projectIndex   int
	projectChanged bool

	// Type and State dropdowns
	typeValues  []model.BundleValue
	typeCursor  int
	stateValues []model.BundleValue
	stateCursor int

	// Assignee autocomplete
	assigneeInput    textinput.Model
	assigneeResults  []model.User
	assigneeCursor   int
	assigneeSelected *model.User
	assigneeGen      int

	// Summary and Description
	summaryInput textinput.Model
	descInput    textarea.Model

	// Comments pane (edit mode)
	comments     []model.Comment
	commentsView viewport.Model

	// Focus management
	focusIndex issueField
	fieldOrder []issueField

	// Loading state for custom fields
	fieldsLoaded bool
}

func NewIssueDialog() IssueDialog {
	si := textinput.New()
	si.Placeholder = "Issue summary"
	si.Prompt = ""
	si.CharLimit = 200

	di := textarea.New()
	di.Placeholder = "Description"
	di.CharLimit = 10000

	ai := textinput.New()
	ai.Placeholder = "Type to search users..."
	ai.Prompt = ""
	ai.CharLimit = 100

	cv := viewport.New(0, 0)

	return IssueDialog{
		summaryInput:  si,
		descInput:     di,
		assigneeInput: ai,
		commentsView:  cv,
	}
}

// OpenCreate activates the dialog in create mode.
func (d *IssueDialog) OpenCreate(projects []model.Project) tea.Cmd {
	d.mode = modeCreate
	d.active = true
	d.submitted = false
	d.issueID = ""

	d.projects = projects
	d.projectIndex = 0
	d.projectChanged = false

	d.typeValues = nil
	d.typeCursor = 0
	d.stateValues = nil
	d.stateCursor = 0
	d.fieldsLoaded = false

	d.assigneeInput.SetValue("")
	d.assigneeResults = nil
	d.assigneeCursor = 0
	d.assigneeSelected = nil
	d.assigneeGen = 0

	d.summaryInput.SetValue("")
	d.descInput.SetValue("")

	d.comments = nil

	d.fieldOrder = []issueField{fieldProject, fieldType, fieldState, fieldAssignee, fieldSummary, fieldDescription}
	d.focusIndex = fieldProject

	return d.updateFocus()
}

// OpenEdit activates the dialog in edit mode with pre-populated values.
func (d *IssueDialog) OpenEdit(issue *model.Issue, comments []model.Comment) tea.Cmd {
	d.mode = modeEdit
	d.active = true
	d.submitted = false
	d.issueID = issue.IDReadable

	d.projects = nil
	d.projectIndex = 0
	d.projectChanged = false

	d.typeValues = nil
	d.typeCursor = 0
	d.stateValues = nil
	d.stateCursor = 0
	d.fieldsLoaded = false

	// Pre-populate assignee
	d.assigneeInput.SetValue("")
	d.assigneeResults = nil
	d.assigneeCursor = 0
	d.assigneeSelected = nil
	d.assigneeGen = 0
	if assignee := issue.AssigneeValue(); assignee != nil {
		d.assigneeSelected = assignee
		d.assigneeInput.SetValue(assignee.Login)
	}

	d.summaryInput.SetValue(issue.Summary)
	d.descInput.SetValue(issue.Description)

	d.comments = comments

	d.fieldOrder = []issueField{fieldType, fieldState, fieldAssignee, fieldSummary, fieldDescription}
	if len(comments) > 0 {
		d.fieldOrder = append(d.fieldOrder, fieldComments)
	}
	d.focusIndex = d.fieldOrder[0]

	return d.updateFocus()
}

// SetCustomFields populates the Type and State dropdowns from fetched project custom fields.
// For edit mode, it also pre-selects the issue's current values.
func (d *IssueDialog) SetCustomFields(fields []model.ProjectCustomField, currentState, currentType string) {
	d.fieldsLoaded = true
	for _, f := range fields {
		switch f.Field.Name {
		case "State":
			d.stateValues = f.Bundle.Values
			d.stateCursor = 0
			if currentState != "" {
				for i, v := range f.Bundle.Values {
					if v.Name == currentState {
						d.stateCursor = i
						break
					}
				}
			}
		case "Type":
			d.typeValues = f.Bundle.Values
			d.typeCursor = 0
			if currentType != "" {
				for i, v := range f.Bundle.Values {
					if v.Name == currentType {
						d.typeCursor = i
						break
					}
				}
			}
		}
	}
}

func (d *IssueDialog) Close() {
	d.active = false
	d.summaryInput.Blur()
	d.descInput.Blur()
	d.assigneeInput.Blur()
}

// focusOrderIndex returns the position of focusIndex within fieldOrder.
func (d *IssueDialog) focusOrderIndex() int {
	for i, f := range d.fieldOrder {
		if f == d.focusIndex {
			return i
		}
	}
	return 0
}

func (d *IssueDialog) nextField() {
	idx := d.focusOrderIndex()
	idx = (idx + 1) % len(d.fieldOrder)
	d.focusIndex = d.fieldOrder[idx]
}

func (d *IssueDialog) prevField() {
	idx := d.focusOrderIndex()
	idx = (idx + len(d.fieldOrder) - 1) % len(d.fieldOrder)
	d.focusIndex = d.fieldOrder[idx]
}

func (d *IssueDialog) updateFocus() tea.Cmd {
	d.summaryInput.Blur()
	d.descInput.Blur()
	d.assigneeInput.Blur()
	switch d.focusIndex {
	case fieldAssignee:
		return d.assigneeInput.Focus()
	case fieldSummary:
		return d.summaryInput.Focus()
	case fieldDescription:
		return d.descInput.Focus()
	}
	return nil
}

// isValid returns true if all mandatory fields have values.
func (d *IssueDialog) isValid() bool {
	if d.mode == modeCreate && len(d.projects) == 0 {
		return false
	}
	if len(d.typeValues) == 0 || len(d.stateValues) == 0 {
		return false
	}
	if d.assigneeSelected == nil {
		return false
	}
	if d.summaryInput.Value() == "" {
		return false
	}
	if d.descInput.Value() == "" {
		return false
	}
	return true
}

// remainingFields returns count of unfilled mandatory fields.
func (d *IssueDialog) remainingFields() int {
	count := 0
	if len(d.typeValues) == 0 {
		count++
	}
	if len(d.stateValues) == 0 {
		count++
	}
	if d.assigneeSelected == nil {
		count++
	}
	if d.summaryInput.Value() == "" {
		count++
	}
	if d.descInput.Value() == "" {
		count++
	}
	return count
}

// buildCustomFields constructs the customFields payload for Create/Update.
func (d *IssueDialog) buildCustomFields() []map[string]any {
	var fields []map[string]any

	if len(d.stateValues) > 0 {
		sv := d.stateValues[d.stateCursor]
		fields = append(fields, map[string]any{
			"name":  "State",
			"$type": "StateIssueCustomField",
			"value": map[string]string{
				"name":  sv.Name,
				"$type": "StateBundleElement",
			},
		})
	}

	if len(d.typeValues) > 0 {
		tv := d.typeValues[d.typeCursor]
		fields = append(fields, map[string]any{
			"name":  "Type",
			"$type": "SingleEnumIssueCustomField",
			"value": map[string]string{
				"name":  tv.Name,
				"$type": "EnumBundleElement",
			},
		})
	}

	if d.assigneeSelected != nil {
		fields = append(fields, map[string]any{
			"name":  "Assignee",
			"$type": "SingleUserIssueCustomField",
			"value": map[string]any{
				"login": d.assigneeSelected.Login,
				"$type": "User",
			},
		})
	}

	return fields
}

// SetAssigneeResults handles autocomplete results with generation guard.
func (d *IssueDialog) SetAssigneeResults(users []model.User, gen int) {
	if gen != d.assigneeGen {
		return // stale results
	}
	d.assigneeResults = users
	d.assigneeCursor = 0
}

func (d *IssueDialog) Update(msg tea.Msg) (IssueDialog, tea.Cmd) {
	if !d.active {
		return *d, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			d.Close()
			return *d, nil

		case "ctrl+s":
			if d.isValid() {
				d.submitted = true
				d.Close()
			}
			return *d, nil

		case "tab":
			d.nextField()
			return *d, d.updateFocus()

		case "shift+tab":
			d.prevField()
			return *d, d.updateFocus()
		}

		// Focus-scoped key handling
		switch d.focusIndex {
		case fieldProject:
			switch msg.String() {
			case "left", "h":
				if len(d.projects) > 0 {
					d.projectIndex = (d.projectIndex + len(d.projects) - 1) % len(d.projects)
					d.projectChanged = true
					d.fieldsLoaded = false
					d.typeValues = nil
					d.stateValues = nil
				}
				return *d, nil
			case "right", "l":
				if len(d.projects) > 0 {
					d.projectIndex = (d.projectIndex + 1) % len(d.projects)
					d.projectChanged = true
					d.fieldsLoaded = false
					d.typeValues = nil
					d.stateValues = nil
				}
				return *d, nil
			}
			return *d, nil

		case fieldType:
			switch msg.String() {
			case "up", "k":
				if d.typeCursor > 0 {
					d.typeCursor--
				}
				return *d, nil
			case "down", "j":
				if d.typeCursor < len(d.typeValues)-1 {
					d.typeCursor++
				}
				return *d, nil
			}
			return *d, nil

		case fieldState:
			switch msg.String() {
			case "up", "k":
				if d.stateCursor > 0 {
					d.stateCursor--
				}
				return *d, nil
			case "down", "j":
				if d.stateCursor < len(d.stateValues)-1 {
					d.stateCursor++
				}
				return *d, nil
			}
			return *d, nil

		case fieldAssignee:
			// If showing autocomplete results and user navigates/selects
			if len(d.assigneeResults) > 0 && d.assigneeSelected == nil {
				switch msg.String() {
				case "up", "k":
					if d.assigneeCursor > 0 {
						d.assigneeCursor--
					}
					return *d, nil
				case "down", "j":
					if d.assigneeCursor < len(d.assigneeResults)-1 {
						d.assigneeCursor++
					}
					return *d, nil
				case "enter":
					user := d.assigneeResults[d.assigneeCursor]
					d.assigneeSelected = &user
					d.assigneeInput.SetValue(user.Login)
					d.assigneeResults = nil
					d.assigneeCursor = 0
					d.nextField()
					return *d, d.updateFocus()
				}
			}

			// Route to text input and debounce
			var cmd tea.Cmd
			d.assigneeInput, cmd = d.assigneeInput.Update(msg)

			// Clear previous selection when typing changes
			if d.assigneeSelected != nil && d.assigneeInput.Value() != d.assigneeSelected.Login {
				d.assigneeSelected = nil
			}

			// Debounce search
			if d.assigneeInput.Value() != "" && d.assigneeSelected == nil {
				d.assigneeGen++
				gen := d.assigneeGen
				debounceCmd := tea.Tick(300*time.Millisecond, func(t time.Time) tea.Msg {
					return assigneeDebounceMsg{generation: gen}
				})
				return *d, tea.Batch(cmd, debounceCmd)
			}
			if d.assigneeInput.Value() == "" {
				d.assigneeResults = nil
			}
			return *d, cmd

		case fieldSummary:
			var cmd tea.Cmd
			d.summaryInput, cmd = d.summaryInput.Update(msg)
			return *d, cmd

		case fieldDescription:
			var cmd tea.Cmd
			d.descInput, cmd = d.descInput.Update(msg)
			return *d, cmd

		case fieldComments:
			switch msg.String() {
			case "up", "k":
				d.commentsView.LineUp(1)
				return *d, nil
			case "down", "j":
				d.commentsView.LineDown(1)
				return *d, nil
			}
			return *d, nil
		}
	}

	// Non-key messages: route to focused text inputs for cursor blink etc.
	switch d.focusIndex {
	case fieldAssignee:
		var cmd tea.Cmd
		d.assigneeInput, cmd = d.assigneeInput.Update(msg)
		return *d, cmd
	case fieldSummary:
		var cmd tea.Cmd
		d.summaryInput, cmd = d.summaryInput.Update(msg)
		return *d, cmd
	case fieldDescription:
		var cmd tea.Cmd
		d.descInput, cmd = d.descInput.Update(msg)
		return *d, cmd
	}

	return *d, nil
}

func (d *IssueDialog) View(width, height int) string {
	if !d.active {
		return ""
	}

	outerWidth := width
	outerHeight := height
	hasComments := len(d.comments) > 0

	// Calculate pane widths
	formWidth := outerWidth - 4 // padding + border
	commentsWidth := 0
	if hasComments {
		formWidth = int(float64(outerWidth)*0.6) - 4
		commentsWidth = outerWidth - int(float64(outerWidth)*0.6) - 4
	}

	formContent := d.renderForm(formWidth, outerHeight-4)

	if !hasComments {
		style := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("99")).
			Padding(1, 1).
			Width(outerWidth - 2).
			Height(outerHeight - 2)
		return style.Render(formContent)
	}

	// Two panes: form left, comments right
	formHeight := outerHeight - 2
	leftPanel := renderTitledPanel(d.formTitle(), formContent, formWidth, formHeight-2, d.focusIndex != fieldComments, lipgloss.Color("99"))
	rightContent := d.renderComments(commentsWidth, formHeight-4)
	rightPanel := renderTitledPanel(iconFile+" Comments", rightContent, commentsWidth, formHeight-2, d.focusIndex == fieldComments, lipgloss.Color("69"))

	return lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel)
}

func (d *IssueDialog) formTitle() string {
	if d.mode == modeCreate {
		return iconFile + " Create Issue"
	}
	return iconFile + " Edit " + d.issueID
}

func (d *IssueDialog) renderForm(width, height int) string {
	var b strings.Builder
	labelFocused := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("69"))
	labelNormal := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	selectedStyle := lipgloss.NewStyle().Background(lipgloss.Color("237")).Foreground(lipgloss.Color("255"))
	cursorStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("69"))

	label := func(name string, field issueField) string {
		if d.focusIndex == field {
			return labelFocused.Render(name + ":")
		}
		return labelNormal.Render(name + ":")
	}

	// Project selector (create only)
	if d.mode == modeCreate {
		b.WriteString(label("Project", fieldProject) + " ")
		if len(d.projects) > 0 {
			p := d.projects[d.projectIndex]
			if d.focusIndex == fieldProject {
				b.WriteString(cursorStyle.Render(fmt.Sprintf("< %s (%s) >", p.Name, p.ShortName)))
			} else {
				b.WriteString(fmt.Sprintf("%s (%s)", p.Name, p.ShortName))
			}
		} else {
			b.WriteString("(no projects)")
		}
		b.WriteString("\n\n")
	}

	// Type dropdown
	b.WriteString(label("Type", fieldType) + "\n")
	if !d.fieldsLoaded {
		b.WriteString("  Loading...\n")
	} else if len(d.typeValues) == 0 {
		b.WriteString("  (none available)\n")
	} else if d.focusIndex == fieldType {
		d.renderInlineList(&b, d.typeValues, d.typeCursor, 5, width, selectedStyle)
	} else {
		b.WriteString("  " + d.typeValues[d.typeCursor].Name + "\n")
	}
	b.WriteString("\n")

	// State dropdown
	b.WriteString(label("State", fieldState) + "\n")
	if !d.fieldsLoaded {
		b.WriteString("  Loading...\n")
	} else if len(d.stateValues) == 0 {
		b.WriteString("  (none available)\n")
	} else if d.focusIndex == fieldState {
		d.renderInlineList(&b, d.stateValues, d.stateCursor, 5, width, selectedStyle)
	} else {
		if d.stateCursor < len(d.stateValues) {
			b.WriteString("  " + stateColor(d.stateValues[d.stateCursor].Name) + "\n")
		}
	}
	b.WriteString("\n")

	// Assignee autocomplete
	b.WriteString(label("Assignee", fieldAssignee) + "\n")
	if d.focusIndex == fieldAssignee {
		d.assigneeInput.Width = width - 4
		b.WriteString("  " + d.assigneeInput.View() + "\n")
		if len(d.assigneeResults) > 0 && d.assigneeSelected == nil {
			for i, u := range d.assigneeResults {
				line := fmt.Sprintf("  %s (%s)", u.Login, u.FullName)
				if len(line) > width {
					line = line[:width-1] + "…"
				}
				if i == d.assigneeCursor {
					b.WriteString(selectedStyle.Render(line) + "\n")
				} else {
					b.WriteString(line + "\n")
				}
			}
		}
	} else if d.assigneeSelected != nil {
		b.WriteString(fmt.Sprintf("  %s (%s)\n", d.assigneeSelected.Login, d.assigneeSelected.FullName))
	} else {
		b.WriteString("  (none)\n")
	}
	b.WriteString("\n")

	// Summary
	b.WriteString(label("Summary", fieldSummary) + "\n")
	d.summaryInput.Width = width - 4
	b.WriteString("  " + d.summaryInput.View() + "\n\n")

	// Description — calculate remaining height for textarea
	usedLines := strings.Count(b.String(), "\n")
	descHeight := height - usedLines - 4
	if descHeight < 3 {
		descHeight = 3
	}
	d.descInput.SetWidth(width - 4)
	d.descInput.SetHeight(descHeight)
	b.WriteString(label("Description", fieldDescription) + "\n")
	b.WriteString("  " + d.descInput.View() + "\n\n")

	// Hint bar
	hint := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	remaining := d.remainingFields()
	hintText := "tab: next  shift+tab: prev  esc: cancel"
	if remaining > 0 {
		hintText = fmt.Sprintf("%d field(s) remaining  |  %s", remaining, hintText)
	} else {
		hintText = "ctrl+s: submit  |  " + hintText
	}
	b.WriteString(hint.Render(hintText))

	return b.String()
}

func (d *IssueDialog) renderInlineList(b *strings.Builder, values []model.BundleValue, cursor, maxVisible, width int, selectedStyle lipgloss.Style) {
	start := 0
	if cursor >= maxVisible {
		start = cursor - maxVisible + 1
	}
	end := start + maxVisible
	if end > len(values) {
		end = len(values)
	}

	if start > 0 {
		b.WriteString("  ▲\n")
	}
	for i := start; i < end; i++ {
		line := "  " + values[i].Name
		if len(line) > width {
			line = line[:width-1] + "…"
		}
		if i == cursor {
			b.WriteString(selectedStyle.Render(line) + "\n")
		} else {
			b.WriteString(line + "\n")
		}
	}
	if end < len(values) {
		b.WriteString("  ▼\n")
	}
}

func (d *IssueDialog) renderComments(width, height int) string {
	var b strings.Builder

	for _, c := range d.comments {
		author := "Unknown"
		if c.Author != nil {
			author = c.Author.FullName
		}
		ts := ""
		if c.Created > 0 {
			ts = " (" + formatTimestamp(c.Created) + ")"
		}
		header := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("69"))
		b.WriteString(header.Render(author+ts) + "\n")
		b.WriteString(c.Text + "\n\n")
	}

	content := b.String()
	d.commentsView.Width = width
	d.commentsView.Height = height
	d.commentsView.SetContent(content)

	return d.commentsView.View()
}
