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

func makeCustomField(name, value string) model.CustomField {
	v := []byte(`{"name": "` + value + `"}`)
	return model.CustomField{Name: name, Value: v}
}

func makeCustomFieldUser(name, login, fullName string) model.CustomField {
	v := []byte(`{"login": "` + login + `", "fullName": "` + fullName + `"}`)
	return model.CustomField{Name: name, Value: v}
}
