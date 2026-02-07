package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/cf/lazytrack/internal/config"
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

func TestNotificationDialog_ViewInactive(t *testing.T) {
	d := NewNotificationDialog()
	result := d.View(80, 24)
	if result != "" {
		t.Errorf("expected empty view when inactive, got %q", result)
	}
}

func TestNotificationDialog_ViewLoading(t *testing.T) {
	d := NewNotificationDialog()
	d.Open(0)

	result := d.View(80, 24)

	if !strings.Contains(result, "Mentions") {
		t.Error("expected title 'Mentions' in view")
	}
	if !strings.Contains(result, "Loading") {
		t.Error("expected loading indicator in view")
	}
}

func TestNotificationDialog_ViewWithResults(t *testing.T) {
	d := NewNotificationDialog()
	d.Open(1000)
	d.SetResults([]model.Issue{
		{IDReadable: "PROJ-1", Summary: "Old issue", Updated: 500},
		{IDReadable: "PROJ-2", Summary: "New issue", Updated: 2000},
	})

	result := d.View(100, 30)

	if !strings.Contains(result, "PROJ-1") {
		t.Error("expected PROJ-1 in view")
	}
	if !strings.Contains(result, "PROJ-2") {
		t.Error("expected PROJ-2 in view")
	}
	if !strings.Contains(result, "NEW") {
		t.Error("expected NEW badge for issue with Updated > lastChecked")
	}
}

func TestNotificationDialog_ViewEmpty(t *testing.T) {
	d := NewNotificationDialog()
	d.Open(0)
	d.SetResults([]model.Issue{})

	result := d.View(80, 24)

	if !strings.Contains(result, "No mentions found") {
		t.Error("expected empty state message")
	}
}

func TestNotificationDialog_ViewError(t *testing.T) {
	d := NewNotificationDialog()
	d.Open(0)
	d.SetError("connection failed")

	result := d.View(80, 24)

	if !strings.Contains(result, "connection failed") {
		t.Error("expected error message in view")
	}
}

func TestLatestIssueTimestamp_Empty(t *testing.T) {
	got := latestIssueTimestamp(nil)
	if got != 0 {
		t.Errorf("got %d, want 0 for empty slice", got)
	}
}

func TestLatestIssueTimestamp_Multiple(t *testing.T) {
	issues := []model.Issue{
		{Updated: 1000},
		{Updated: 3000},
		{Updated: 2000},
	}
	got := latestIssueTimestamp(issues)
	if got != 3000 {
		t.Errorf("got %d, want 3000", got)
	}
}

func TestFetchMentionsCmd_NoProject(t *testing.T) {
	svc := &mentionRecordingService{}
	app := NewApp(svc, config.DefaultState())

	cmd := app.fetchMentionsCmd()
	if cmd == nil {
		t.Fatal("expected non-nil cmd")
	}
	msg := cmd()
	if _, ok := msg.(mentionsLoadedMsg); !ok {
		t.Errorf("expected mentionsLoadedMsg, got %T", msg)
	}
	if svc.lastQuery == "" {
		t.Fatal("expected ListIssues to be called")
	}
	if !strings.Contains(svc.lastQuery, "mentioned: me") {
		t.Errorf("query %q should contain 'mentioned: me'", svc.lastQuery)
	}
}

func TestFetchMentionsCmd_WithProject(t *testing.T) {
	svc := &mentionRecordingService{}
	app := NewApp(svc, config.DefaultState())
	app.activeProject = &model.Project{ShortName: "PROJ"}

	cmd := app.fetchMentionsCmd()
	msg := cmd()
	if _, ok := msg.(mentionsLoadedMsg); !ok {
		t.Errorf("expected mentionsLoadedMsg, got %T", msg)
	}
	if !strings.Contains(svc.lastQuery, "project: PROJ") {
		t.Errorf("query %q should contain 'project: PROJ'", svc.lastQuery)
	}
	if !strings.Contains(svc.lastQuery, "mentioned: me") {
		t.Errorf("query %q should contain 'mentioned: me'", svc.lastQuery)
	}
}

// mentionRecordingService records ListIssues calls for mention query tests.
type mentionRecordingService struct {
	mockService
	lastQuery string
}

func (m *mentionRecordingService) ListIssues(query string, skip, top int) ([]model.Issue, error) {
	m.lastQuery = query
	return []model.Issue{}, nil
}

// --- App-level integration tests ---

func TestApp_NKey_OpensNotificationDialog(t *testing.T) {
	svc := &mentionRecordingService{}
	app := NewApp(svc, config.DefaultState())
	app.ready = true
	app.width = 120
	app.height = 40
	app.currentUser = &model.User{Login: "testuser"}

	m, cmd := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	a := m.(*App)

	if !a.notifDialog.active {
		t.Error("expected notification dialog to be active")
	}
	if cmd == nil {
		t.Fatal("expected non-nil cmd to fetch mentions")
	}
}

func TestApp_NKey_NoUser(t *testing.T) {
	app := NewApp(&mockService{}, config.DefaultState())
	app.ready = true
	app.width = 120
	app.height = 40
	// currentUser is nil

	m, _ := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	a := m.(*App)

	if a.notifDialog.active {
		t.Error("notification dialog should not open when currentUser is nil")
	}
	if a.err == "" {
		t.Error("expected error message when currentUser is nil")
	}
}

func TestApp_MentionsLoadedMsg_CountsUnread(t *testing.T) {
	app := NewApp(&mockService{}, config.DefaultState())
	app.ready = true
	app.width = 120
	app.height = 40
	app.lastCheckedMentions = 2000

	issues := []model.Issue{
		{IDReadable: "A-1", Updated: 1000}, // old
		{IDReadable: "A-2", Updated: 3000}, // new
		{IDReadable: "A-3", Updated: 5000}, // new
	}
	m, _ := app.Update(mentionsLoadedMsg{issues})
	a := m.(*App)

	if a.unreadMentionCount != 2 {
		t.Errorf("got unreadMentionCount %d, want 2", a.unreadMentionCount)
	}
	if len(a.mentionedIssues) != 3 {
		t.Errorf("got %d mentionedIssues, want 3", len(a.mentionedIssues))
	}
}

func TestApp_MentionsLoadedMsg_PopulatesOpenDialog(t *testing.T) {
	app := NewApp(&mockService{}, config.DefaultState())
	app.ready = true
	app.width = 120
	app.height = 40
	app.notifDialog.Open(0) // dialog is open and loading

	issues := []model.Issue{
		{IDReadable: "A-1", Summary: "First"},
	}
	m, _ := app.Update(mentionsLoadedMsg{issues})
	a := m.(*App)

	if a.notifDialog.loading {
		t.Error("expected dialog loading to be false after mentionsLoadedMsg")
	}
	if len(a.notifDialog.issues) != 1 {
		t.Errorf("got %d dialog issues, want 1", len(a.notifDialog.issues))
	}
}

func TestApp_NotifDialog_EnterNavigatesToIssue(t *testing.T) {
	svc := &recordingService{}
	app := NewApp(svc, config.DefaultState())
	app.ready = true
	app.width = 120
	app.height = 40
	app.lastCheckedMentions = 0
	app.mentionedIssues = []model.Issue{
		{IDReadable: "PROJ-42", Summary: "Test", Updated: 5000},
	}

	// Open dialog and set results
	app.notifDialog.Open(0)
	app.notifDialog.SetResults(app.mentionedIssues)

	// Press enter to select
	m, cmd := app.Update(tea.KeyMsg{Type: tea.KeyEnter})
	a := m.(*App)

	if a.notifDialog.active {
		t.Error("expected dialog to close after enter")
	}
	if a.lastCheckedMentions != 5000 {
		t.Errorf("got lastCheckedMentions %d, want 5000", a.lastCheckedMentions)
	}
	if a.unreadMentionCount != 0 {
		t.Errorf("got unreadMentionCount %d, want 0", a.unreadMentionCount)
	}
	if !a.listCollapsed {
		t.Error("expected list collapsed after navigation")
	}
	if a.focus != detailPane {
		t.Error("expected focus on detail pane")
	}
	if cmd == nil {
		t.Fatal("expected non-nil cmd to fetch issue detail")
	}
}

func TestApp_NotifDialog_EscDoesNotUpdateTimestamp(t *testing.T) {
	app := NewApp(&mockService{}, config.DefaultState())
	app.ready = true
	app.width = 120
	app.height = 40
	app.lastCheckedMentions = 1000
	app.mentionedIssues = []model.Issue{
		{IDReadable: "A-1", Updated: 5000},
	}

	app.notifDialog.Open(1000)
	app.notifDialog.SetResults(app.mentionedIssues)

	// Press esc
	m, _ := app.Update(tea.KeyMsg{Type: tea.KeyEsc})
	a := m.(*App)

	if a.notifDialog.active {
		t.Error("expected dialog to close after esc")
	}
	if a.lastCheckedMentions != 1000 {
		t.Errorf("got lastCheckedMentions %d, want 1000 (unchanged)", a.lastCheckedMentions)
	}
}

func TestApp_Refresh_IncludesMentions(t *testing.T) {
	svc := &mentionRecordingService{}
	app := NewApp(svc, config.DefaultState())
	app.ready = true
	app.width = 120
	app.height = 40
	app.currentUser = &model.User{Login: "testuser"}

	m, cmd := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	a := m.(*App)

	if !a.loading {
		t.Error("expected loading after refresh")
	}
	if cmd == nil {
		t.Fatal("expected non-nil cmd")
	}

	// Execute the batch to trigger service calls
	msg := cmd()
	if batch, ok := msg.(tea.BatchMsg); ok {
		for _, fn := range batch {
			fn()
		}
	}

	if svc.lastQuery == "" {
		t.Error("expected ListIssues to be called for mentions during refresh")
	}
}
