package ui

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

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
			"$type": original.StateFieldType(),
			"value": map[string]string{
				"name":  parsed.state,
				"$type": "StateBundleElement",
			},
		})
	}

	if parsed.issueType != original.TypeValue() {
		customFields = append(customFields, map[string]any{
			"name":  "Type",
			"$type": original.TypeFieldType(),
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
