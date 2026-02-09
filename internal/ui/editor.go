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
