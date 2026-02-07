package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

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

func TestNotificationDialog_EscCloses(t *testing.T) {
	d := NewNotificationDialog()
	d.Open(0)
	d.SetResults([]model.Issue{{IDReadable: "A-1"}})

	d, _ = d.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if d.active {
		t.Error("expected not active after esc")
	}
	if d.submitted {
		t.Error("expected not submitted after esc")
	}
}

func TestNotificationDialog_EnterSelectsIssue(t *testing.T) {
	d := NewNotificationDialog()
	d.Open(0)
	d.SetResults([]model.Issue{
		{IDReadable: "A-1", Summary: "First"},
		{IDReadable: "A-2", Summary: "Second"},
	})

	// Move to second item
	d, _ = d.Update(tea.KeyMsg{Type: tea.KeyDown})
	// Select it
	d, _ = d.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if !d.submitted {
		t.Error("expected submitted after enter")
	}
	if d.selectedIssue == nil {
		t.Fatal("expected non-nil selectedIssue")
	}
	if d.selectedIssue.IDReadable != "A-2" {
		t.Errorf("got selectedIssue %q, want A-2", d.selectedIssue.IDReadable)
	}
	if d.active {
		t.Error("expected not active after enter")
	}
}

func TestNotificationDialog_EnterNoResults(t *testing.T) {
	d := NewNotificationDialog()
	d.Open(0)
	d.SetResults([]model.Issue{})

	d, _ = d.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if d.submitted {
		t.Error("expected not submitted with no results")
	}
}

func TestNotificationDialog_CursorNavigation(t *testing.T) {
	d := NewNotificationDialog()
	d.Open(0)
	d.SetResults([]model.Issue{
		{IDReadable: "A-1"},
		{IDReadable: "A-2"},
		{IDReadable: "A-3"},
	})

	// down
	d, _ = d.Update(tea.KeyMsg{Type: tea.KeyDown})
	if d.cursor != 1 {
		t.Errorf("got cursor %d, want 1 after down", d.cursor)
	}

	// j
	d, _ = d.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if d.cursor != 2 {
		t.Errorf("got cursor %d, want 2 after j", d.cursor)
	}

	// Clamp at bottom
	d, _ = d.Update(tea.KeyMsg{Type: tea.KeyDown})
	if d.cursor != 2 {
		t.Errorf("got cursor %d, want 2 (clamped at bottom)", d.cursor)
	}

	// up
	d, _ = d.Update(tea.KeyMsg{Type: tea.KeyUp})
	if d.cursor != 1 {
		t.Errorf("got cursor %d, want 1 after up", d.cursor)
	}

	// k
	d, _ = d.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if d.cursor != 0 {
		t.Errorf("got cursor %d, want 0 after k", d.cursor)
	}

	// Clamp at top
	d, _ = d.Update(tea.KeyMsg{Type: tea.KeyUp})
	if d.cursor != 0 {
		t.Errorf("got cursor %d, want 0 (clamped at top)", d.cursor)
	}
}

func TestNotificationDialog_InactiveUpdate(t *testing.T) {
	d := NewNotificationDialog()
	// Not opened — should be a no-op
	d, cmd := d.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if d.active {
		t.Error("should remain inactive")
	}
	if cmd != nil {
		t.Error("expected nil cmd when inactive")
	}
}
