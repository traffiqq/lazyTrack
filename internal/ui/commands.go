package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// effectiveQuery returns the search query with project filter prepended if applicable.
func (a *App) effectiveQuery() string {
	query := a.query
	if a.activeProject != nil && !strings.Contains(strings.ToLower(query), "project:") {
		query = "project: " + a.activeProject.ShortName + " " + query
		query = strings.TrimSpace(query)
	}
	return query
}

// resolveGotoProject returns the ShortName of the project to use for goto.
// Priority: activeProject > selected issue > first loaded issue.
// Returns "" if no project context is available.
func (a *App) resolveGotoProject() string {
	if a.activeProject != nil {
		return a.activeProject.ShortName
	}
	if a.selected != nil && a.selected.Project != nil {
		return a.selected.Project.ShortName
	}
	if len(a.issues) > 0 && a.issues[0].Project != nil {
		return a.issues[0].Project.ShortName
	}
	return ""
}

// fetchIssuesCmd creates a command that fetches issues. Captures current query value.
func (a *App) fetchIssuesCmd() tea.Cmd {
	query := a.effectiveQuery()
	service := a.service
	return func() tea.Msg {
		issues, err := service.ListIssues(query, 0, 50)
		if err != nil {
			return errMsg{err}
		}
		return issuesLoadedMsg{issues}
	}
}

// fetchMoreIssuesCmd creates a command to load the next page of issues.
func (a *App) fetchMoreIssuesCmd() tea.Cmd {
	query := a.query
	skip := len(a.issues)
	pageSize := a.pageSize
	service := a.service
	return func() tea.Msg {
		issues, err := service.ListIssues(query, skip, pageSize)
		if err != nil {
			return errMsg{err}
		}
		return moreIssuesLoadedMsg{issues}
	}
}

// fetchDetailCmd creates a command that fetches issue detail. Captures issueID.
func (a *App) fetchDetailCmd(issueID string) tea.Cmd {
	service := a.service
	return func() tea.Msg {
		issue, err := service.GetIssue(issueID)
		if err != nil {
			return errMsg{err}
		}
		return issueDetailLoadedMsg{issue}
	}
}
