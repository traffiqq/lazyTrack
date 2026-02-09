# Nvim Edit Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Allow editing an issue in `$EDITOR` (nvim) with YAML front matter for metadata and markdown body for description.

**Architecture:** Write issue to temp file, suspend TUI via `tea.ExecProcess`, user edits in nvim, parse file back on return, diff against original, update via API. All editor logic in a new `editor.go` file.

**Tech Stack:** Go, Bubble Tea (`tea.ExecProcess`), `os/exec`, `os`

---

### Task 1: Create editor.go with resolveEditor and temp file writing

**Files:**
- Create: `internal/ui/editor.go`
- Test: `internal/ui/editor_test.go`

**Step 1: Write the tests**

Create `internal/ui/editor_test.go`:

```go
package ui

import (
	"os"
	"strings"
	"testing"

	"github.com/cf/lazytrack/internal/model"
)

func TestResolveEditor(t *testing.T) {
	// Save and restore original EDITOR
	orig := os.Getenv("EDITOR")
	defer os.Setenv("EDITOR", orig)

	os.Setenv("EDITOR", "nano")
	if got := resolveEditor(); got != "nano" {
		t.Errorf("resolveEditor() = %q, want %q", got, "nano")
	}

	os.Setenv("EDITOR", "")
	got := resolveEditor()
	// Should fall back to nvim, vim, or vi (whichever is on PATH)
	if got != "nvim" && got != "vim" && got != "vi" {
		t.Errorf("resolveEditor() = %q, want nvim/vim/vi", got)
	}
}

func TestWriteIssueTempFile(t *testing.T) {
	issue := &model.Issue{
		IDReadable:  "TEST-1",
		Summary:     "Fix login bug",
		Description: "The login page throws a 500 error.",
	}
	// Add custom fields for state, assignee, type
	issue.CustomFields = []model.CustomField{
		makeCustomField("State", "In Progress"),
		makeCustomField("Type", "Bug"),
		makeCustomFieldUser("Assignee", "johndoe", "John Doe"),
	}

	path, err := writeIssueTempFile(issue)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(path)

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)

	if !strings.Contains(content, "summary: Fix login bug") {
		t.Error("missing summary in temp file")
	}
	if !strings.Contains(content, "state: In Progress") {
		t.Error("missing state in temp file")
	}
	if !strings.Contains(content, "assignee: johndoe") {
		t.Error("missing assignee in temp file")
	}
	if !strings.Contains(content, "type: Bug") {
		t.Error("missing type in temp file")
	}
	if !strings.Contains(content, "The login page throws a 500 error.") {
		t.Error("missing description in temp file")
	}
	// Verify front matter delimiters
	if !strings.HasPrefix(content, "---\n") {
		t.Error("temp file should start with ---")
	}
	if strings.Count(content, "---\n") < 2 {
		t.Error("temp file should have two --- delimiters")
	}
}

func TestWriteIssueTempFile_EmptyFields(t *testing.T) {
	issue := &model.Issue{
		IDReadable:  "TEST-2",
		Summary:     "Minimal issue",
		Description: "",
	}

	path, err := writeIssueTempFile(issue)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(path)

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)

	if !strings.Contains(content, "assignee:") {
		t.Error("should have empty assignee field")
	}
	if !strings.Contains(content, "state:") {
		t.Error("should have empty state field")
	}
}
```

**Step 2: Write test helpers for custom fields**

Add these helpers at the bottom of `internal/ui/editor_test.go`:

```go
func makeCustomField(name, value string) model.CustomField {
	v := []byte(`{"name": "` + value + `"}`)
	return model.CustomField{Name: name, Value: v}
}

func makeCustomFieldUser(name, login, fullName string) model.CustomField {
	v := []byte(`{"login": "` + login + `", "fullName": "` + fullName + `"}`)
	return model.CustomField{Name: name, Value: v}
}
```

**Step 3: Run tests to verify they fail**

Run: `go test ./internal/ui/ -v -run "TestResolveEditor|TestWriteIssueTempFile"`
Expected: FAIL (functions don't exist yet)

**Step 4: Implement editor.go**

Create `internal/ui/editor.go`:

```go
package ui

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/cf/lazytrack/internal/model"
)

// resolveEditor returns the editor command to use.
// Checks $EDITOR, then falls back to nvim, vim, vi.
func resolveEditor() string {
	if editor := os.Getenv("EDITOR"); editor != "" {
		return editor
	}
	for _, name := range []string{"nvim", "vim", "vi"} {
		if _, err := exec.LookPath(name); err == nil {
			return name
		}
	}
	return "vi"
}

// writeIssueTempFile writes an issue to a temp file in front matter + body format.
// Returns the temp file path.
func writeIssueTempFile(issue *model.Issue) (string, error) {
	assignee := ""
	if u := issue.AssigneeValue(); u != nil {
		assignee = u.Login
	}

	content := fmt.Sprintf(`---
summary: %s
state: %s
assignee: %s
type: %s
---

%s
`, issue.Summary, issue.StateValue(), assignee, issue.TypeValue(), issue.Description)

	f, err := os.CreateTemp("", fmt.Sprintf("lazytrack-edit-%s-*.md", issue.IDReadable))
	if err != nil {
		return "", err
	}
	defer f.Close()

	if _, err := f.WriteString(content); err != nil {
		os.Remove(f.Name())
		return "", err
	}

	return f.Name(), nil
}
```

**Step 5: Run tests to verify they pass**

Run: `go test ./internal/ui/ -v -run "TestResolveEditor|TestWriteIssueTempFile"`
Expected: PASS

**Step 6: Commit**

```bash
git add internal/ui/editor.go internal/ui/editor_test.go
git commit -m "feat: add resolveEditor and writeIssueTempFile"
```

---

### Task 2: Add parsing and diffing functions

**Files:**
- Modify: `internal/ui/editor.go`
- Modify: `internal/ui/editor_test.go`

**Step 1: Write the tests**

Add to `internal/ui/editor_test.go`:

```go
func TestParseIssueTempFile(t *testing.T) {
	content := `---
summary: Updated summary
state: Fixed
assignee: janedoe
type: Task
---

Updated description here.
`
	f, err := os.CreateTemp("", "lazytrack-test-*.md")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.WriteString(content)
	f.Close()

	parsed, err := parseIssueTempFile(f.Name())
	if err != nil {
		t.Fatal(err)
	}

	if parsed.summary != "Updated summary" {
		t.Errorf("summary = %q, want %q", parsed.summary, "Updated summary")
	}
	if parsed.state != "Fixed" {
		t.Errorf("state = %q, want %q", parsed.state, "Fixed")
	}
	if parsed.assignee != "janedoe" {
		t.Errorf("assignee = %q, want %q", parsed.assignee, "janedoe")
	}
	if parsed.issueType != "Task" {
		t.Errorf("type = %q, want %q", parsed.issueType, "Task")
	}
	if strings.TrimSpace(parsed.description) != "Updated description here." {
		t.Errorf("description = %q, want %q", strings.TrimSpace(parsed.description), "Updated description here.")
	}
}

func TestParseIssueTempFile_EmptyAssignee(t *testing.T) {
	content := `---
summary: Test
state: Open
assignee:
type: Bug
---

Desc.
`
	f, err := os.CreateTemp("", "lazytrack-test-*.md")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.WriteString(content)
	f.Close()

	parsed, err := parseIssueTempFile(f.Name())
	if err != nil {
		t.Fatal(err)
	}

	if parsed.assignee != "" {
		t.Errorf("assignee = %q, want empty", parsed.assignee)
	}
}

func TestParseIssueTempFile_Malformed(t *testing.T) {
	content := `no front matter here`
	f, err := os.CreateTemp("", "lazytrack-test-*.md")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.WriteString(content)
	f.Close()

	_, err = parseIssueTempFile(f.Name())
	if err == nil {
		t.Error("expected error for malformed file")
	}
}

func TestBuildEditorUpdateFields_NoChanges(t *testing.T) {
	issue := &model.Issue{
		Summary:     "Original",
		Description: "Desc",
		CustomFields: []model.CustomField{
			makeCustomField("State", "Open"),
			makeCustomField("Type", "Bug"),
			makeCustomFieldUser("Assignee", "alice", "Alice"),
		},
	}
	parsed := parsedIssue{
		summary:     "Original",
		description: "Desc",
		state:       "Open",
		issueType:   "Bug",
		assignee:    "alice",
	}

	fields := buildEditorUpdateFields(issue, parsed)
	if fields != nil {
		t.Errorf("expected nil fields for no changes, got %v", fields)
	}
}

func TestBuildEditorUpdateFields_SummaryChanged(t *testing.T) {
	issue := &model.Issue{
		Summary:     "Original",
		Description: "Desc",
		CustomFields: []model.CustomField{
			makeCustomField("State", "Open"),
			makeCustomField("Type", "Bug"),
		},
	}
	parsed := parsedIssue{
		summary:     "Updated title",
		description: "Desc",
		state:       "Open",
		issueType:   "Bug",
	}

	fields := buildEditorUpdateFields(issue, parsed)
	if fields == nil {
		t.Fatal("expected non-nil fields")
	}
	if fields["summary"] != "Updated title" {
		t.Errorf("summary = %v, want %q", fields["summary"], "Updated title")
	}
}

func TestBuildEditorUpdateFields_AssigneeCleared(t *testing.T) {
	issue := &model.Issue{
		Summary:     "Test",
		Description: "Desc",
		CustomFields: []model.CustomField{
			makeCustomField("State", "Open"),
			makeCustomField("Type", "Bug"),
			makeCustomFieldUser("Assignee", "alice", "Alice"),
		},
	}
	parsed := parsedIssue{
		summary:     "Test",
		description: "Desc",
		state:       "Open",
		issueType:   "Bug",
		assignee:    "", // cleared
	}

	fields := buildEditorUpdateFields(issue, parsed)
	if fields == nil {
		t.Fatal("expected non-nil fields")
	}
	cf, ok := fields["customFields"]
	if !ok {
		t.Fatal("expected customFields key")
	}
	cfSlice := cf.([]map[string]any)
	found := false
	for _, f := range cfSlice {
		if f["name"] == "Assignee" && f["value"] == nil {
			found = true
		}
	}
	if !found {
		t.Error("expected Assignee with nil value for unassign")
	}
}

func TestBuildEditorUpdateFields_StateChanged(t *testing.T) {
	issue := &model.Issue{
		Summary:     "Test",
		Description: "Desc",
		CustomFields: []model.CustomField{
			makeCustomField("State", "Open"),
			makeCustomField("Type", "Bug"),
		},
	}
	parsed := parsedIssue{
		summary:     "Test",
		description: "Desc",
		state:       "Fixed",
		issueType:   "Bug",
	}

	fields := buildEditorUpdateFields(issue, parsed)
	if fields == nil {
		t.Fatal("expected non-nil fields")
	}
	cf := fields["customFields"].([]map[string]any)
	found := false
	for _, f := range cf {
		if f["name"] == "State" {
			val := f["value"].(map[string]string)
			if val["name"] == "Fixed" {
				found = true
			}
		}
	}
	if !found {
		t.Error("expected State custom field set to Fixed")
	}
}
```

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/ui/ -v -run "TestParse|TestBuildEditor"`
Expected: FAIL (types/functions don't exist yet)

**Step 3: Implement parsing and diffing in editor.go**

Add to `internal/ui/editor.go`:

```go
// parsedIssue holds the values extracted from an edited temp file.
type parsedIssue struct {
	summary     string
	state       string
	assignee    string
	issueType   string
	description string
}

// parseIssueTempFile reads a temp file and extracts front matter + description.
func parseIssueTempFile(path string) (parsedIssue, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return parsedIssue{}, err
	}

	content := string(data)
	parts := strings.SplitN(content, "---\n", 3)
	if len(parts) < 3 {
		return parsedIssue{}, fmt.Errorf("malformed file: missing front matter delimiters")
	}

	var p parsedIssue
	for _, line := range strings.Split(parts[1], "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		key, val, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		switch strings.TrimSpace(key) {
		case "summary":
			p.summary = strings.TrimSpace(val)
		case "state":
			p.state = strings.TrimSpace(val)
		case "assignee":
			p.assignee = strings.TrimSpace(val)
		case "type":
			p.issueType = strings.TrimSpace(val)
		}
	}

	p.description = strings.TrimSpace(parts[2])
	return p, nil
}

// buildEditorUpdateFields compares parsed values against the original issue
// and returns a fields map for UpdateIssue. Returns nil if nothing changed.
func buildEditorUpdateFields(original *model.Issue, parsed parsedIssue) map[string]any {
	fields := map[string]any{}
	var customFields []map[string]any

	if parsed.summary != original.Summary {
		fields["summary"] = parsed.summary
	}

	if parsed.description != original.Description {
		fields["description"] = parsed.description
	}

	if parsed.state != original.StateValue() {
		customFields = append(customFields, map[string]any{
			"name":  "State",
			"$type": "StateIssueCustomField",
			"value": map[string]string{
				"name":  parsed.state,
				"$type": "StateBundleElement",
			},
		})
	}

	if parsed.issueType != original.TypeValue() {
		customFields = append(customFields, map[string]any{
			"name":  "Type",
			"$type": "EnumIssueCustomField",
			"value": map[string]string{
				"name":  parsed.issueType,
				"$type": "EnumBundleElement",
			},
		})
	}

	origAssignee := ""
	if u := original.AssigneeValue(); u != nil {
		origAssignee = u.Login
	}
	if parsed.assignee != origAssignee {
		if parsed.assignee == "" {
			customFields = append(customFields, map[string]any{
				"name":  "Assignee",
				"$type": "SingleUserIssueCustomField",
				"value": nil,
			})
		} else {
			customFields = append(customFields, map[string]any{
				"name":  "Assignee",
				"$type": "SingleUserIssueCustomField",
				"value": map[string]any{
					"login": parsed.assignee,
					"$type": "User",
				},
			})
		}
	}

	if len(customFields) > 0 {
		fields["customFields"] = customFields
	}

	if len(fields) == 0 {
		return nil
	}
	return fields
}
```

Also add `"strings"` to the import block in `editor.go` (alongside the existing imports).

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/ui/ -v -run "TestParse|TestBuildEditor"`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/ui/editor.go internal/ui/editor_test.go
git commit -m "feat: add parseIssueTempFile and buildEditorUpdateFields"
```

---

### Task 3: Add message type and wire into leader dispatch + Update

**Files:**
- Modify: `internal/ui/messages.go`
- Modify: `internal/ui/keyhandling.go:155-252`
- Modify: `internal/ui/app.go:181-396`

**Step 1: Add editorFinishedMsg to messages.go**

Add after line 67 in `internal/ui/messages.go`:

```go
type editorFinishedMsg struct {
	err      error
	tempPath string
	original *model.Issue
}
```

Also add `"github.com/cf/lazytrack/internal/model"` import — it is already imported, so no change needed.

**Step 2: Add `case "v":` to leader dispatch in keyhandling.go**

In `internal/ui/keyhandling.go`, inside the leader dispatch `switch msg.String()` block (around line 248, before the closing `}`), add:

```go
		case "v":
			if a.selected != nil {
				issue := a.selected
				tempPath, err := writeIssueTempFile(issue)
				if err != nil {
					a.err = "Failed to create temp file: " + err.Error()
					return a, nil
				}
				editor := resolveEditor()
				c := exec.Command(editor, tempPath)
				return a, tea.ExecProcess(c, func(err error) tea.Msg {
					return editorFinishedMsg{err: err, tempPath: tempPath, original: issue}
				})
			}
```

Also add `"os/exec"` to the imports in `keyhandling.go`.

**Step 3: Handle editorFinishedMsg in app.go Update**

In `internal/ui/app.go`, inside the `Update` method's type switch, add a new case. Place it before the `case tea.KeyMsg:` block (around line 391):

```go
	case editorFinishedMsg:
		if msg.tempPath != "" {
			defer os.Remove(msg.tempPath)
		}
		if msg.err != nil {
			a.err = "Editor error: " + msg.err.Error()
			return a, nil
		}
		parsed, err := parseIssueTempFile(msg.tempPath)
		if err != nil {
			a.err = "Parse error: " + err.Error()
			return a, nil
		}
		fields := buildEditorUpdateFields(msg.original, parsed)
		if fields == nil {
			return a, nil // nothing changed
		}
		issueID := msg.original.IDReadable
		service := a.service
		a.loading = true
		return a, func() tea.Msg {
			err := service.UpdateIssue(issueID, fields)
			if err != nil {
				return errMsg{err}
			}
			return issueUpdatedMsg{}
		}
```

Also add `"os"` to the imports in `app.go`.

**Step 4: Verify it compiles**

Run: `go build ./...`
Expected: success

**Step 5: Run full test suite**

Run: `go test ./... -v`
Expected: all pass

**Step 6: Commit**

```bash
git add internal/ui/messages.go internal/ui/keyhandling.go internal/ui/app.go
git commit -m "feat: wire editor into leader dispatch and handle editorFinishedMsg"
```

---

### Task 4: Update leader hints, help overlay, and status bar

**Files:**
- Modify: `internal/ui/styles.go:84-95`
- Modify: `internal/ui/help.go:20-30`

**Step 1: Add vim edit to leaderHints**

In `internal/ui/styles.go`, in the `leaderHints` slice (line 84-95), add `{"v", "vim edit"}` in alphabetical order — after `{"t", "toggle"}`:

```go
	leaderHints = []keyHint{
		{"a", "assign"},
		{"c", "create"},
		{"d", "delete"},
		{"e", "edit"},
		{"f", "find"},
		{"m", "comment"},
		{"n", "notifs"},
		{"p", "project"},
		{"s", "state"},
		{"t", "toggle"},
		{"v", "vim edit"},
	}
```

**Step 2: Add vim edit to help text**

In `internal/ui/help.go`, add `space v     Vim edit issue` after `space t     Toggle issue list` (line 30):

```
  space t     Toggle issue list
  space v     Vim edit issue
```

**Step 3: Verify it compiles**

Run: `go build ./...`
Expected: success

**Step 4: Commit**

```bash
git add internal/ui/styles.go internal/ui/help.go
git commit -m "feat: add vim edit to leader hints and help overlay"
```

---

### Task 5: Run full tests and verify

**Step 1: Run full test suite**

Run: `go test ./... -v`
Expected: all pass

**Step 2: Run linter**

Run: `go vet ./...`
Expected: no issues

**Step 3: Build binary**

Run: `make build`
Expected: `lazytrack` binary built successfully
