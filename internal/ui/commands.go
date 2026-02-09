package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/cf/lazytrack/internal/model"
)

// effectiveQuery returns the search query with project filter, filter bar clauses,
// and user query combined.
func (a *App) effectiveQuery() string {
	var parts []string

	if a.activeProject != nil && !strings.Contains(strings.ToLower(a.query), "project:") {
		parts = append(parts, "project: "+a.activeProject.ShortName)
	}

	if a.filterMe {
		parts = append(parts, "Assignee: me")
	}

	if tf := buildTypeFilter(a.filterBug, a.filterTask); tf != "" {
		parts = append(parts, tf)
	}

	if a.query != "" {
		parts = append(parts, a.query)
	}

	return strings.Join(parts, " ")
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
	query := a.effectiveQuery()
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

// fetchCurrentUserCmd creates a command that fetches the current user.
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

// fetchMentionsCmd creates a command that fetches issues mentioning the current user.
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

// buildTypeFilter returns a YouTrack Type filter clause.
// Returns "Type: Bug" or "Type: Task" when one is set,
// "Type: Bug,Task" when both are set, or "" when neither is set.
func buildTypeFilter(bug, task bool) string {
	switch {
	case bug && task:
		return "Type: Bug,Task"
	case bug:
		return "Type: Bug"
	case task:
		return "Type: Task"
	default:
		return ""
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
