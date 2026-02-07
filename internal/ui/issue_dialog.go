package ui

import (
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

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
	typeValues    []model.BundleValue
	typeCursor    int
	typeFieldType string // e.g. "SingleEnumIssueCustomField"
	stateValues   []model.BundleValue
	stateCursor   int
	stateFieldType string // e.g. "StateIssueCustomField"

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
			d.stateFieldType = f.Field.Type
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
			d.typeFieldType = f.Field.Type
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
			"$type": d.stateFieldType,
			"value": map[string]string{
				"name":  sv.Name,
				"$type": sv.Type,
			},
		})
	}

	if len(d.typeValues) > 0 {
		tv := d.typeValues[d.typeCursor]
		fields = append(fields, map[string]any{
			"name":  "Type",
			"$type": d.typeFieldType,
			"value": map[string]string{
				"name":  tv.Name,
				"$type": tv.Type,
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
