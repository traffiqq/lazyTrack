package ui

import (
	"testing"

	"github.com/cf/lazytrack/internal/model"
)

func TestNotificationDialog_OpenSetsActive(t *testing.T) {
	d := NewNotificationDialog()
	d.Open(1000)

	if !d.active {
		t.Error("expected active after Open")
	}
	if !d.loading {
		t.Error("expected loading after Open")
	}
	if d.lastChecked != 1000 {
		t.Errorf("got lastChecked %d, want 1000", d.lastChecked)
	}
}

func TestNotificationDialog_OpenResets(t *testing.T) {
	d := NewNotificationDialog()
	// Simulate prior state
	d.submitted = true
	d.cursor = 5
	d.err = "old error"
	issue := model.Issue{IDReadable: "X-1"}
	d.selectedIssue = &issue

	d.Open(2000)

	if d.submitted {
		t.Error("expected submitted reset to false")
	}
	if d.cursor != 0 {
		t.Errorf("got cursor %d, want 0", d.cursor)
	}
	if d.err != "" {
		t.Errorf("got err %q, want empty", d.err)
	}
	if d.selectedIssue != nil {
		t.Error("expected selectedIssue reset to nil")
	}
}

func TestNotificationDialog_SetResults(t *testing.T) {
	d := NewNotificationDialog()
	d.Open(0)

	issues := []model.Issue{
		{IDReadable: "A-1", Summary: "First"},
		{IDReadable: "A-2", Summary: "Second"},
	}
	d.SetResults(issues)

	if d.loading {
		t.Error("expected loading false after SetResults")
	}
	if len(d.issues) != 2 {
		t.Errorf("got %d issues, want 2", len(d.issues))
	}
}

func TestNotificationDialog_SetError(t *testing.T) {
	d := NewNotificationDialog()
	d.Open(0)

	d.SetError("fetch failed")

	if d.loading {
		t.Error("expected loading false after SetError")
	}
	if d.err != "fetch failed" {
		t.Errorf("got err %q, want %q", d.err, "fetch failed")
	}
}

func TestNotificationDialog_Close(t *testing.T) {
	d := NewNotificationDialog()
	d.Open(0)
	d.Close()

	if d.active {
		t.Error("expected not active after Close")
	}
}
