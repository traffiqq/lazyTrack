package ui

import (
	"testing"

	"github.com/cf/lazytrack/internal/config"
	"github.com/cf/lazytrack/internal/model"
)

func TestEffectiveQuery_NoActiveProject(t *testing.T) {
	app := NewApp(&mockService{}, config.DefaultState())
	app.query = "sort by: updated desc"

	got := app.effectiveQuery()
	if got != "sort by: updated desc" {
		t.Errorf("got %q, want %q", got, "sort by: updated desc")
	}
}

func TestEffectiveQuery_WithActiveProject(t *testing.T) {
	app := NewApp(&mockService{}, config.DefaultState())
	app.activeProject = &model.Project{ShortName: "PROJ"}
	app.query = "#Unresolved"

	got := app.effectiveQuery()
	want := "project: PROJ #Unresolved"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestEffectiveQuery_WithActiveProject_EmptyQuery(t *testing.T) {
	app := NewApp(&mockService{}, config.DefaultState())
	app.activeProject = &model.Project{ShortName: "PROJ"}
	app.query = ""

	got := app.effectiveQuery()
	want := "project: PROJ"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestEffectiveQuery_ManualProjectOverride(t *testing.T) {
	app := NewApp(&mockService{}, config.DefaultState())
	app.activeProject = &model.Project{ShortName: "PROJ"}
	app.query = "project: OTHER #Unresolved"

	got := app.effectiveQuery()
	want := "project: OTHER #Unresolved"
	if got != want {
		t.Errorf("got %q, want %q — active project should be skipped when query contains project:", got, want)
	}
}

func TestEffectiveQuery_CaseInsensitiveProjectDetection(t *testing.T) {
	app := NewApp(&mockService{}, config.DefaultState())
	app.activeProject = &model.Project{ShortName: "PROJ"}
	app.query = "Project: OTHER"

	got := app.effectiveQuery()
	want := "Project: OTHER"
	if got != want {
		t.Errorf("got %q, want %q — case-insensitive project: detection", got, want)
	}
}

func TestResolveGotoProject_ActiveProject(t *testing.T) {
	app := NewApp(&mockService{}, config.DefaultState())
	app.activeProject = &model.Project{ShortName: "ACTIVE"}

	got := app.resolveGotoProject()
	if got != "ACTIVE" {
		t.Errorf("got %q, want ACTIVE", got)
	}
}

func TestResolveGotoProject_FallbackToSelected(t *testing.T) {
	app := NewApp(&mockService{}, config.DefaultState())
	app.selected = &model.Issue{
		Project: &model.Project{ShortName: "SEL"},
	}

	got := app.resolveGotoProject()
	if got != "SEL" {
		t.Errorf("got %q, want SEL", got)
	}
}

func TestResolveGotoProject_FallbackToFirstIssue(t *testing.T) {
	app := NewApp(&mockService{}, config.DefaultState())
	app.issues = []model.Issue{
		{Project: &model.Project{ShortName: "FIRST"}},
	}

	got := app.resolveGotoProject()
	if got != "FIRST" {
		t.Errorf("got %q, want FIRST", got)
	}
}

func TestResolveGotoProject_NoContext(t *testing.T) {
	app := NewApp(&mockService{}, config.DefaultState())

	got := app.resolveGotoProject()
	if got != "" {
		t.Errorf("got %q, want empty", got)
	}
}

// mockService implements IssueService for testing.
type mockService struct{}

func (m *mockService) GetCurrentUser() (*model.User, error)                          { return nil, nil }
func (m *mockService) ListIssues(query string, skip, top int) ([]model.Issue, error) { return nil, nil }
func (m *mockService) GetIssue(issueID string) (*model.Issue, error)                 { return nil, nil }
func (m *mockService) CreateIssue(projectID, summary, description string, customFields []map[string]any) (*model.Issue, error) {
	return nil, nil
}
func (m *mockService) UpdateIssue(issueID string, fields map[string]any) error { return nil }
func (m *mockService) DeleteIssue(issueID string) error                        { return nil }
func (m *mockService) ListComments(issueID string) ([]model.Comment, error)    { return nil, nil }
func (m *mockService) AddComment(issueID, text string) (*model.Comment, error) { return nil, nil }
func (m *mockService) ListProjects() ([]model.Project, error)                  { return nil, nil }
func (m *mockService) SearchUsers(query string) ([]model.User, error)          { return nil, nil }
func (m *mockService) ListProjectCustomFields(projectID string) ([]model.ProjectCustomField, error) {
	return nil, nil
}
