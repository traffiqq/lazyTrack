package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/cf/lazytrack/internal/model"
)

func renderIssueDetail(issue *model.Issue, width int) string {
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
		b.WriteString(renderMarkdown(issue.Description, width) + "\n")
	} else {
		b.WriteString("(no description)\n")
	}

	return b.String()
}

func renderComments(comments []model.Comment, width int) string {
	var b strings.Builder

	for i := len(comments) - 1; i >= 0; i-- {
		c := comments[i]
		author := "Unknown"
		if c.Author != nil {
			if c.Author.FullName != "" {
				author = c.Author.FullName
			} else if c.Author.Login != "" {
				author = c.Author.Login
			}
		}
		ts := ""
		if c.Created > 0 {
			ts = " (" + formatTimestamp(c.Created) + ")"
		}
		fmt.Fprintf(&b, "%s %s%s\n%s\n", iconComment, author, ts, renderMarkdown(c.Text, width))
		if i > 0 {
			b.WriteString("\n")
		}
	}

	return b.String()
}

func formatTimestamp(ms int64) string {
	t := time.UnixMilli(ms)
	return t.Format("2006-01-02 15:04")
}
