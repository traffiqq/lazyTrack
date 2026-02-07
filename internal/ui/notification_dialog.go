package ui

import (
	tea "github.com/charmbracelet/bubbletea"

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
