package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/cf/lazytrack/internal/config"
	"github.com/cf/lazytrack/internal/model"
)

// recordingService wraps mockService and records which methods are called.
type recordingService struct {
	mockService
	listIssuesCalls int
	getIssueCalls   []string
}

func (r *recordingService) ListIssues(query string, skip, top int) ([]model.Issue, error) {
	r.listIssuesCalls++
	return nil, nil
}

func (r *recordingService) GetIssue(issueID string) (*model.Issue, error) {
	r.getIssueCalls = append(r.getIssueCalls, issueID)
	return &model.Issue{IDReadable: issueID}, nil
}

func TestRefresh_NoSelectedIssue(t *testing.T) {
	svc := &recordingService{}
	app := NewApp(svc, config.DefaultState())
	app.ready = true
	app.width = 120
	app.height = 40

	// Press "r" with no selected issue
	m, cmd := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	a := m.(*App)

	if !a.loading {
		t.Error("expected loading to be true after refresh")
	}
	if cmd == nil {
		t.Fatal("expected non-nil cmd")
	}

	// Execute the batched command to trigger service calls
	msg := cmd()
	// The batch returns a tea.BatchMsg containing sub-commands
	if batch, ok := msg.(tea.BatchMsg); ok {
		for _, fn := range batch {
			fn()
		}
	}

	if svc.listIssuesCalls != 1 {
		t.Errorf("expected 1 ListIssues call, got %d", svc.listIssuesCalls)
	}
	if len(svc.getIssueCalls) != 0 {
		t.Errorf("expected 0 GetIssue calls, got %d", len(svc.getIssueCalls))
	}
}

func TestRefresh_WithSelectedIssue(t *testing.T) {
	svc := &recordingService{}
	app := NewApp(svc, config.DefaultState())
	app.ready = true
	app.width = 120
	app.height = 40
	app.selected = &model.Issue{IDReadable: "PROJ-42"}

	// Press "r" with a selected issue
	m, cmd := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	a := m.(*App)

	if !a.loading {
		t.Error("expected loading to be true after refresh")
	}
	if cmd == nil {
		t.Fatal("expected non-nil cmd")
	}

	// Execute the batched command to trigger service calls
	msg := cmd()
	if batch, ok := msg.(tea.BatchMsg); ok {
		for _, fn := range batch {
			fn()
		}
	}

	if svc.listIssuesCalls != 1 {
		t.Errorf("expected 1 ListIssues call, got %d", svc.listIssuesCalls)
	}
	if len(svc.getIssueCalls) != 1 {
		t.Errorf("expected 1 GetIssue call, got %d", len(svc.getIssueCalls))
	}
	if len(svc.getIssueCalls) > 0 && svc.getIssueCalls[0] != "PROJ-42" {
		t.Errorf("expected GetIssue(PROJ-42), got GetIssue(%s)", svc.getIssueCalls[0])
	}
}

func TestRefresh_BlockedDuringSearch(t *testing.T) {
	svc := &recordingService{}
	app := NewApp(svc, config.DefaultState())
	app.ready = true
	app.width = 120
	app.height = 40
	app.searching = true

	// Press "r" while searching — should be consumed by search input, not trigger refresh
	m, _ := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	a := m.(*App)

	if a.loading {
		t.Error("expected loading to be false — refresh should not trigger during search")
	}
	if svc.listIssuesCalls != 0 {
		t.Errorf("expected 0 ListIssues calls during search, got %d", svc.listIssuesCalls)
	}
}
