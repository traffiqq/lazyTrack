package ui

import "github.com/cf/lazytrack/internal/model"

// IssueService defines the operations the UI needs from the API layer.
type IssueService interface {
	GetCurrentUser() (*model.User, error)
	ListIssues(query string, skip, top int) ([]model.Issue, error)
	GetIssue(issueID string) (*model.Issue, error)
	CreateIssue(projectID, summary, description string, customFields []map[string]any) (*model.Issue, error)
	UpdateIssue(issueID string, fields map[string]any) error
	DeleteIssue(issueID string) error
	ListComments(issueID string) ([]model.Comment, error)
	AddComment(issueID, text string) (*model.Comment, error)
	ListProjects() ([]model.Project, error)
	SearchUsers(query string) ([]model.User, error)
	ListProjectCustomFields(projectID string) ([]model.ProjectCustomField, error)
}
