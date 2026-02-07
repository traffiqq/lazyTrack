package ui

import (
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
