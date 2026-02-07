package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/cf/lazytrack/internal/model"
)

func TestProjectPicker_OpenSetsActive(t *testing.T) {
	d := NewProjectPickerDialog()
	projects := []model.Project{
		{ID: "1", Name: "Alpha", ShortName: "ALPHA"},
		{ID: "2", Name: "Beta", ShortName: "BETA"},
	}
	d.Open(projects)

	if !d.active {
		t.Error("expected active after Open")
	}
	// cursor starts at 0 which is "(All Projects)"
	if d.cursor != 0 {
		t.Errorf("got cursor %d, want 0", d.cursor)
	}
	// total entries: 1 (All Projects) + 2 projects = 3
	if len(d.projects) != 2 {
		t.Errorf("got %d projects, want 2", len(d.projects))
	}
}

func TestProjectPicker_SelectAllProjects(t *testing.T) {
	d := NewProjectPickerDialog()
	projects := []model.Project{
		{ID: "1", Name: "Alpha", ShortName: "ALPHA"},
	}
	d.Open(projects)

	// cursor is at 0 = "(All Projects)", press enter
	d, _ = d.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if !d.submitted {
		t.Error("expected submitted after enter")
	}
	if d.selectedProject != nil {
		t.Error("expected nil selectedProject for All Projects")
	}
}

func TestProjectPicker_SelectSpecificProject(t *testing.T) {
	d := NewProjectPickerDialog()
	projects := []model.Project{
		{ID: "1", Name: "Alpha", ShortName: "ALPHA"},
		{ID: "2", Name: "Beta", ShortName: "BETA"},
	}
	d.Open(projects)

	// Move down to first project (index 1)
	d, _ = d.Update(tea.KeyMsg{Type: tea.KeyDown})
	if d.cursor != 1 {
		t.Errorf("got cursor %d, want 1", d.cursor)
	}

	// Press enter
	d, _ = d.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if !d.submitted {
		t.Error("expected submitted")
	}
	if d.selectedProject == nil {
		t.Fatal("expected non-nil selectedProject")
	}
	if d.selectedProject.ShortName != "ALPHA" {
		t.Errorf("got ShortName %q, want ALPHA", d.selectedProject.ShortName)
	}
}

func TestProjectPicker_EscCloses(t *testing.T) {
	d := NewProjectPickerDialog()
	d.Open([]model.Project{{ID: "1", Name: "A", ShortName: "A"}})

	d, _ = d.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if d.active {
		t.Error("expected not active after esc")
	}
	if d.submitted {
		t.Error("expected not submitted after esc")
	}
}

func TestProjectPicker_CursorBounds(t *testing.T) {
	d := NewProjectPickerDialog()
	projects := []model.Project{
		{ID: "1", Name: "Alpha", ShortName: "ALPHA"},
	}
	d.Open(projects)
	// total items: 2 (All Projects + Alpha)

	// Try to go above 0
	d, _ = d.Update(tea.KeyMsg{Type: tea.KeyUp})
	if d.cursor != 0 {
		t.Errorf("got cursor %d, want 0 (clamped)", d.cursor)
	}

	// Go down to 1
	d, _ = d.Update(tea.KeyMsg{Type: tea.KeyDown})
	if d.cursor != 1 {
		t.Errorf("got cursor %d, want 1", d.cursor)
	}

	// Try to go past end
	d, _ = d.Update(tea.KeyMsg{Type: tea.KeyDown})
	if d.cursor != 1 {
		t.Errorf("got cursor %d, want 1 (clamped)", d.cursor)
	}
}

func TestProjectPicker_JKNavigation(t *testing.T) {
	d := NewProjectPickerDialog()
	d.Open([]model.Project{
		{ID: "1", Name: "A", ShortName: "A"},
		{ID: "2", Name: "B", ShortName: "B"},
	})

	d, _ = d.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if d.cursor != 1 {
		t.Errorf("got cursor %d, want 1 after j", d.cursor)
	}

	d, _ = d.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if d.cursor != 0 {
		t.Errorf("got cursor %d, want 0 after k", d.cursor)
	}
}
