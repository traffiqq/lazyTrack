# Notifications / Mentions Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add a mentions notification system with a status bar badge, overlay dialog, and read/unread tracking.

**Architecture:** Reuse `ListIssues` with `mentioned: me` query. New `NotificationDialog` component follows the `FinderDialog`/`ProjectPickerDialog` pattern (struct with `active`/`submitted`/`selectedIssue`, `Update()`/`View()` methods, manual scroll math). Unread count is timestamp-based, persisted in `state.yaml`.

**Tech Stack:** Go, Bubble Tea, Lip Gloss, existing `IssueService` interface.

---

### Task 1: State Persistence — `LastCheckedMentions`

**Files:**
- Modify: `internal/config/state.go:12-21` (UIState struct)
- Modify: `internal/config/state_test.go`

**Step 1: Write the failing test**

Add to `internal/config/state_test.go`:

```go
func TestSaveState_LastCheckedMentions_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "state.yaml")

	state := State{
		UI: UIState{
			ListRatio:           0.4,
			LastCheckedMentions: 1707300000000,
		},
	}

	if err := SaveStateToPath(path, state); err != nil {
		t.Fatalf("save error: %v", err)
	}

	loaded := LoadStateFromPath(path)

	if loaded.UI.LastCheckedMentions != 1707300000000 {
		t.Errorf("got LastCheckedMentions %d, want 1707300000000", loaded.UI.LastCheckedMentions)
	}
}

func TestLoadState_LastCheckedMentions_DefaultsToZero(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "state.yaml")

	content := []byte(`ui:
  list_ratio: 0.4
`)
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatal(err)
	}

	state := LoadStateFromPath(path)

	if state.UI.LastCheckedMentions != 0 {
		t.Errorf("got LastCheckedMentions %d, want 0", state.UI.LastCheckedMentions)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/config/ -v -run TestSaveState_LastCheckedMentions_RoundTrip`
Expected: FAIL — `LastCheckedMentions` field does not exist.

**Step 3: Write minimal implementation**

In `internal/config/state.go`, add the field to the `UIState` struct:

```go
type UIState struct {
	ListRatio           float64 `yaml:"list_ratio"`
	ListCollapsed       bool    `yaml:"list_collapsed"`
	SelectedIssue       string  `yaml:"selected_issue"`
	ActiveProject       string  `yaml:"active_project,omitempty"`
	LastCheckedMentions int64   `yaml:"last_checked_mentions,omitempty"`
}
```

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/config/ -v`
Expected: ALL PASS

**Step 5: Commit**

```bash
git add internal/config/state.go internal/config/state_test.go
git commit -m "feat(config): add LastCheckedMentions to UIState"
```

---

### Task 2: NotificationDialog — Struct, Constructor, Lifecycle

**Files:**
- Create: `internal/ui/notification_dialog.go`
- Create: `internal/ui/notification_dialog_test.go`

**Step 1: Write the failing tests**

Create `internal/ui/notification_dialog_test.go`:

```go
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
```

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/ui/ -v -run TestNotificationDialog`
Expected: FAIL — types not defined.

**Step 3: Write minimal implementation**

Create `internal/ui/notification_dialog.go`:

```go
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
```

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/ui/ -v -run TestNotificationDialog`
Expected: ALL PASS

**Step 5: Commit**

```bash
git add internal/ui/notification_dialog.go internal/ui/notification_dialog_test.go
git commit -m "feat(ui): add NotificationDialog struct, constructor, lifecycle"
```

---

### Task 3: NotificationDialog — Update (Key Handling)

**Files:**
- Modify: `internal/ui/notification_dialog.go`
- Modify: `internal/ui/notification_dialog_test.go`

**Step 1: Write the failing tests**

Append to `internal/ui/notification_dialog_test.go`:

```go
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
```

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/ui/ -v -run TestNotificationDialog_Esc`
Expected: FAIL — `Update` method not defined.

**Step 3: Write minimal implementation**

Add to `internal/ui/notification_dialog.go`:

```go
import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/cf/lazytrack/internal/model"
)

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
```

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/ui/ -v -run TestNotificationDialog`
Expected: ALL PASS

**Step 5: Commit**

```bash
git add internal/ui/notification_dialog.go internal/ui/notification_dialog_test.go
git commit -m "feat(ui): add NotificationDialog Update with key handling"
```

---

### Task 4: NotificationDialog — View

**Files:**
- Modify: `internal/ui/notification_dialog.go`
- Modify: `internal/ui/notification_dialog_test.go`

**Step 1: Write the failing tests**

Append to `internal/ui/notification_dialog_test.go`:

```go
import "strings"

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
```

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/ui/ -v -run TestNotificationDialog_View`
Expected: FAIL — `View` method not defined.

**Step 3: Write minimal implementation**

Add to `internal/ui/notification_dialog.go`:

```go
import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/cf/lazytrack/internal/model"
)

func (d *NotificationDialog) View(width, height int) string {
	if !d.active {
		return ""
	}

	dialogWidth := width * 3 / 5
	if dialogWidth < 50 {
		dialogWidth = 50
	}
	dialogHeight := height * 7 / 10
	if dialogHeight < 15 {
		dialogHeight = 15
	}
	if dialogHeight > height-2 {
		dialogHeight = height - 2
	}

	contentWidth := dialogWidth - 6

	var b strings.Builder

	b.WriteString(titleStyle.Render("Mentions") + "\n\n")

	resultsHeight := dialogHeight - 7
	if resultsHeight < 3 {
		resultsHeight = 3
	}

	if d.err != "" {
		b.WriteString(errorStyle.Render("Error: "+d.err) + "\n")
	} else if d.loading {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("Loading mentions...") + "\n")
	} else if len(d.issues) == 0 {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("No mentions found") + "\n")
	} else {
		normalStyle := lipgloss.NewStyle().Width(contentWidth)
		selectedStyle := lipgloss.NewStyle().
			Width(contentWidth).
			Background(lipgloss.Color("237")).
			Foreground(lipgloss.Color("255"))
		newBadge := lipgloss.NewStyle().
			Foreground(lipgloss.Color("220")).
			Bold(true)

		start := 0
		if d.cursor >= resultsHeight {
			start = d.cursor - resultsHeight + 1
		}
		end := start + resultsHeight
		if end > len(d.issues) {
			end = len(d.issues)
		}

		for i := start; i < end; i++ {
			issue := d.issues[i]
			badge := "     "
			if issue.Updated > d.lastChecked {
				badge = newBadge.Render("[NEW]")
			}
			line := fmt.Sprintf("%s %-12s %s", badge, issue.IDReadable, issue.Summary)
			if lipgloss.Width(line) > contentWidth {
				// Truncate summary to fit
				maxSummary := contentWidth - lipgloss.Width(badge) - 14
				if maxSummary < 0 {
					maxSummary = 0
				}
				summary := issue.Summary
				if len(summary) > maxSummary {
					summary = summary[:maxSummary-1] + "…"
				}
				line = fmt.Sprintf("%s %-12s %s", badge, issue.IDReadable, summary)
			}
			if i == d.cursor {
				b.WriteString(selectedStyle.Render(line) + "\n")
			} else {
				b.WriteString(normalStyle.Render(line) + "\n")
			}
		}
	}

	b.WriteString("\n")
	hint := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	b.WriteString(hint.Render("j/k: navigate  enter: open  esc: close"))

	dialogStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("99")).
		Padding(1, 2).
		Width(dialogWidth).
		Height(dialogHeight)

	dialog := dialogStyle.Render(b.String())

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, dialog)
}
```

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/ui/ -v -run TestNotificationDialog`
Expected: ALL PASS

**Step 5: Commit**

```bash
git add internal/ui/notification_dialog.go internal/ui/notification_dialog_test.go
git commit -m "feat(ui): add NotificationDialog View with NEW badges and scroll"
```

---

### Task 5: Message Types and Commands

**Files:**
- Modify: `internal/ui/messages.go`
- Modify: `internal/ui/commands.go`

**Step 1: Add message types**

Append to `internal/ui/messages.go`:

```go
type currentUserLoadedMsg struct {
	user *model.User
}

type mentionsLoadedMsg struct {
	issues []model.Issue
}
```

**Step 2: Add commands and helper**

Append to `internal/ui/commands.go`:

```go
func (a *App) fetchCurrentUserCmd() tea.Cmd {
	service := a.service
	return func() tea.Msg {
		user, err := service.GetCurrentUser()
		if err != nil {
			return errMsg{err}
		}
		return currentUserLoadedMsg{user}
	}
}

func (a *App) fetchMentionsCmd() tea.Cmd {
	query := "mentioned: me sort by: updated desc"
	if a.activeProject != nil {
		query = "project: " + a.activeProject.ShortName + " " + query
	}
	service := a.service
	return func() tea.Msg {
		issues, err := service.ListIssues(query, 0, 50)
		if err != nil {
			return errMsg{err}
		}
		return mentionsLoadedMsg{issues}
	}
}

// latestIssueTimestamp returns the maximum Updated timestamp from a slice of issues.
// Returns 0 if the slice is empty.
func latestIssueTimestamp(issues []model.Issue) int64 {
	var max int64
	for _, issue := range issues {
		if issue.Updated > max {
			max = issue.Updated
		}
	}
	return max
}
```

**Step 3: Run build to verify it compiles**

Run: `go build ./...`
Expected: SUCCESS (no errors)

**Step 4: Commit**

```bash
git add internal/ui/messages.go internal/ui/commands.go
git commit -m "feat(ui): add mention message types, fetch commands, and timestamp helper"
```

---

### Task 6: Test `latestIssueTimestamp` and `fetchMentionsCmd`

**Files:**
- Modify: `internal/ui/notification_dialog_test.go`

**Step 1: Write the tests**

Append to `internal/ui/notification_dialog_test.go`:

```go
import (
	"github.com/cf/lazytrack/internal/config"
)

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
```

**Step 2: Run tests**

Run: `go test ./internal/ui/ -v -run "TestLatestIssueTimestamp|TestFetchMentionsCmd"`
Expected: ALL PASS

**Step 3: Commit**

```bash
git add internal/ui/notification_dialog_test.go
git commit -m "test(ui): add tests for latestIssueTimestamp and fetchMentionsCmd"
```

---

### Task 7: Mention Badge Style

**Files:**
- Modify: `internal/ui/styles.go`

**Step 1: Add the style**

Add to the second `var` block in `internal/ui/styles.go` (the one with `keyStyle` and `hintDescStyle`), after `hintDescStyle`:

```go
	mentionBadgeStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("220")).
		Bold(true)
```

**Step 2: Run build**

Run: `go build ./...`
Expected: SUCCESS

**Step 3: Commit**

```bash
git add internal/ui/styles.go
git commit -m "feat(ui): add mentionBadgeStyle for status bar badge"
```

---

### Task 8: App Wiring — Fields, Init, Update Handlers

**Files:**
- Modify: `internal/ui/app.go`

**Step 1: Add new fields to `App` struct**

In `internal/ui/app.go`, add after the `gotoProject` field (around line 88):

```go
	notifDialog        NotificationDialog
	currentUser        *model.User
	lastCheckedMentions int64
	mentionedIssues    []model.Issue
	unreadMentionCount int
```

**Step 2: Initialize in `NewApp`**

In the `NewApp` function, add to the `app := &App{...}` block:

```go
		notifDialog:         NewNotificationDialog(),
		lastCheckedMentions: state.UI.LastCheckedMentions,
```

**Step 3: Update `Init()` to fetch current user in parallel**

Change `Init()` from:

```go
func (a *App) Init() tea.Cmd {
	return a.fetchIssuesCmd()
}
```

to:

```go
func (a *App) Init() tea.Cmd {
	return tea.Batch(a.fetchIssuesCmd(), a.fetchCurrentUserCmd())
}
```

**Step 4: Add `currentUserLoadedMsg` handler in `Update()`**

In the `switch msg := msg.(type)` block, add before the `tea.KeyMsg` case:

```go
	case currentUserLoadedMsg:
		a.currentUser = msg.user
		return a, a.fetchMentionsCmd()

	case mentionsLoadedMsg:
		a.mentionedIssues = msg.issues
		a.unreadMentionCount = 0
		for _, issue := range msg.issues {
			if issue.Updated > a.lastCheckedMentions {
				a.unreadMentionCount++
			}
		}
		if a.notifDialog.active && a.notifDialog.loading {
			a.notifDialog.SetResults(msg.issues)
		}
		return a, nil
```

**Step 5: Extend `errMsg` handler**

In the existing `errMsg` handler, add a branch for the notification dialog before the finder dialog branch:

```go
	case errMsg:
		a.loading = false
		if a.notifDialog.active {
			a.notifDialog.SetError(msg.err.Error())
			return a, nil
		}
		if a.finderDialog.active {
```

**Step 6: Run build**

Run: `go build ./...`
Expected: SUCCESS

**Step 7: Run all existing tests to check nothing broke**

Run: `go test ./... -v`
Expected: ALL PASS

**Step 8: Commit**

```bash
git add internal/ui/app.go
git commit -m "feat(ui): wire NotificationDialog into App with init and update handlers"
```

---

### Task 9: Key Handling — Dialog Routing and `n` Key

**Files:**
- Modify: `internal/ui/keyhandling.go`

**Step 1: Add notification dialog routing**

In `handleKeyMsg`, add BEFORE the `if a.finderDialog.active {` block (currently at line 73):

```go
	// When notification dialog is active, route input to it
	if a.notifDialog.active {
		var cmd tea.Cmd
		a.notifDialog, cmd = a.notifDialog.Update(msg)
		if a.notifDialog.submitted && a.notifDialog.selectedIssue != nil {
			issueID := a.notifDialog.selectedIssue.IDReadable
			a.lastCheckedMentions = latestIssueTimestamp(a.mentionedIssues)
			a.unreadMentionCount = 0
			a.listCollapsed = true
			a.focus = detailPane
			a.resizePanels()
			a.loading = true
			return a, a.fetchDetailCmd(issueID)
		}
		return a, cmd
	}
```

**Step 2: Add `n` key binding**

In the global key switch (after the `"f"` case, around line 436):

```go
	case "n":
		if a.currentUser == nil {
			a.err = "Could not load user — mentions unavailable"
			return a, nil
		}
		a.notifDialog.Open(a.lastCheckedMentions)
		service := a.service
		query := "mentioned: me sort by: updated desc"
		if a.activeProject != nil {
			query = "project: " + a.activeProject.ShortName + " " + query
		}
		return a, func() tea.Msg {
			issues, err := service.ListIssues(query, 0, 50)
			if err != nil {
				return errMsg{err}
			}
			return mentionsLoadedMsg{issues}
		}
```

**Step 3: Update refresh (`r`) to include mentions**

In the `"r"` case, add before `return a, tea.Batch(refreshCmds...)`:

```go
		if a.currentUser != nil {
			refreshCmds = append(refreshCmds, a.fetchMentionsCmd())
		}
```

**Step 4: Update quit handler to persist `LastCheckedMentions`**

In the `"ctrl+c", "q"` case, add `LastCheckedMentions` to the state struct:

```go
		state := config.State{
			UI: config.UIState{
				ListRatio:           a.listRatio,
				ListCollapsed:       a.listCollapsed,
				LastCheckedMentions: a.lastCheckedMentions,
			},
		}
```

**Step 5: Run build and tests**

Run: `go build ./... && go test ./... -v`
Expected: ALL PASS

**Step 6: Commit**

```bash
git add internal/ui/keyhandling.go
git commit -m "feat(ui): add notification dialog routing, n key, refresh and quit integration"
```

---

### Task 10: View and Status Bar Integration

**Files:**
- Modify: `internal/ui/view.go`
- Modify: `internal/ui/statusbar.go`

**Step 1: Add notification dialog to overlay chain in `view.go`**

In `View()`, add BEFORE the `if a.finderDialog.active {` line (line 15):

```go
	if a.notifDialog.active {
		return a.notifDialog.View(a.width, a.height)
	}
```

**Step 2: Add mention badge to status bar**

In `renderStatusBar()` in `statusbar.go`, add after the `if a.loading {` block (after line 28) and before the right-side hints section:

```go
	if a.unreadMentionCount > 0 {
		left += mentionBadgeStyle.Render(fmt.Sprintf(" · %d mentions", a.unreadMentionCount))
	}
```

Note: `fmt` is already imported in `statusbar.go` — check and add if needed.

**Step 3: Add `fmt` import to `statusbar.go` if missing**

Check the imports in `statusbar.go`. Currently it imports `"strings"` and `"github.com/charmbracelet/lipgloss"`. Add `"fmt"` to the import block.

**Step 4: Run build and tests**

Run: `go build ./... && go test ./... -v`
Expected: ALL PASS

**Step 5: Commit**

```bash
git add internal/ui/view.go internal/ui/statusbar.go
git commit -m "feat(ui): add notification dialog overlay and mention badge in status bar"
```

---

### Task 11: Help Overlay Update

**Files:**
- Modify: `internal/ui/help.go`

**Step 1: Add `n` keybinding to help text**

In `internal/ui/help.go`, add `n           Mentions` to the Actions section, after the `r           Refresh` line:

```
  n           Mentions
```

**Step 2: Run build**

Run: `go build ./...`
Expected: SUCCESS

**Step 3: Commit**

```bash
git add internal/ui/help.go
git commit -m "feat(ui): add n key to help overlay"
```

---

### Task 12: Integration Tests

**Files:**
- Modify: `internal/ui/notification_dialog_test.go`

**Step 1: Write app-level integration tests**

Append to `internal/ui/notification_dialog_test.go`:

```go
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

	// Should have called ListIssues at least twice (issues + mentions)
	// The mentionRecordingService only records the last query,
	// so just verify it was called with the mention query
	if svc.lastQuery == "" {
		t.Error("expected ListIssues to be called for mentions during refresh")
	}
}
```

**Step 2: Run all tests**

Run: `go test ./... -v`
Expected: ALL PASS

**Step 3: Commit**

```bash
git add internal/ui/notification_dialog_test.go
git commit -m "test(ui): add integration tests for notification dialog wiring"
```

---

### Task 13: Final Verification

**Step 1: Run full test suite**

Run: `go test ./... -v`
Expected: ALL PASS

**Step 2: Run linter**

Run: `go vet ./...`
Expected: No issues

**Step 3: Build the binary**

Run: `go build -o lazytrack ./cmd/lazytrack`
Expected: Binary builds successfully

**Step 4: Verify binary runs (smoke test)**

Run: `./lazytrack --help` or just verify it starts without panic (ctrl+c to exit).

**Step 5: Commit (if any final fixes needed)**

Only if fixes were needed in previous steps.
