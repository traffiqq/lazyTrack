package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/cf/lazytrack/internal/model"
)

func renderIssueDetail(issue *model.Issue) string {
	var b strings.Builder

	b.WriteString(titleStyle.Render(issue.IDReadable+" "+issue.Summary) + "\n\n")

	if issue.Project != nil {
		fmt.Fprintf(&b, "Project: %s (%s)\n", issue.Project.Name, issue.Project.ShortName)
	}

	state := issue.StateValue()
	if state != "" {
		fmt.Fprintf(&b, "State: %s\n", stateColor(state))
	}

	if assignee := issue.AssigneeValue(); assignee != nil {
		fmt.Fprintf(&b, "Assignee: %s\n", assignee.FullName)
	}

	if issue.Reporter != nil {
		fmt.Fprintf(&b, "Reporter: %s\n", issue.Reporter.FullName)
	}

	if issue.Created > 0 {
		fmt.Fprintf(&b, "Created: %s\n", formatTimestamp(issue.Created))
	}
	if issue.Updated > 0 {
		fmt.Fprintf(&b, "Updated: %s\n", formatTimestamp(issue.Updated))
	}

	b.WriteString("\n────────────────────────────────\n\n")

	if issue.Description != "" {
		b.WriteString(issue.Description + "\n")
	} else {
		b.WriteString("(no description)\n")
	}

	if len(issue.Comments) > 0 {
		b.WriteString("\n────────── Comments ──────────\n\n")
		for _, c := range issue.Comments {
			author := "Unknown"
			if c.Author != nil {
				author = c.Author.FullName
			}
			ts := ""
			if c.Created > 0 {
				ts = " (" + formatTimestamp(c.Created) + ")"
			}
			fmt.Fprintf(&b, "%s%s:\n%s\n\n", author, ts, c.Text)
		}
	}

	return b.String()
}

func formatTimestamp(ms int64) string {
	t := time.UnixMilli(ms)
	return t.Format("2006-01-02 15:04")
}
