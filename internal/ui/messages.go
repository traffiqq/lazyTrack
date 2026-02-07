package ui

import "github.com/cf/lazytrack/internal/model"

// Messages used across the TUI.

type issuesLoadedMsg struct {
	issues []model.Issue
}

type moreIssuesLoadedMsg struct {
	issues []model.Issue
}

type issueDetailLoadedMsg struct {
	issue *model.Issue
}

type projectsLoadedMsg struct {
	projects []model.Project
}

type projectsForPickerMsg struct {
	projects []model.Project
}

type issueCreatedMsg struct{}

type issueUpdatedMsg struct{}

type issueDeletedMsg struct{}

type commentAddedMsg struct{}

type errMsg struct {
	err error
}

type finderDebounceMsg struct {
	generation int
}

type finderSearchResultsMsg struct {
	issues     []model.Issue
	generation int
}

type assigneeDebounceMsg struct {
	generation int
}

type assigneeSearchResultsMsg struct {
	users      []model.User
	generation int
}

type customFieldsLoadedMsg struct {
	fields []model.ProjectCustomField
}

type currentUserLoadedMsg struct {
	user *model.User
}

type mentionsLoadedMsg struct {
	issues []model.Issue
}
