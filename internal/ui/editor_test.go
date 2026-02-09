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

func makeCustomField(name, value string) model.CustomField {
	v := []byte(`{"name": "` + value + `"}`)
	return model.CustomField{Name: name, Value: v}
}

func makeCustomFieldUser(name, login, fullName string) model.CustomField {
	v := []byte(`{"login": "` + login + `", "fullName": "` + fullName + `"}`)
	return model.CustomField{Name: name, Value: v}
}
