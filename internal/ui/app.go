package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/cf/lazytrack/internal/config"
	"github.com/cf/lazytrack/internal/model"
)

type pane int

const (
	listPane pane = iota
	detailPane
)

// issueItem wraps model.Issue for the list.Model interface.
type issueItem struct {
	issue model.Issue
}

func (i issueItem) Title() string {
	return fmt.Sprintf("[%s] %s", i.issue.IDReadable, i.issue.Summary)
}

func (i issueItem) Description() string {
	var parts []string
	if t := i.issue.TypeValue(); t != "" {
		parts = append(parts, t)
	}
	if s := i.issue.StateValue(); s != "" {
		parts = append(parts, stateColor(s))
	}
	if a := i.issue.AssigneeValue(); a != nil {
		name := a.FullName
		if name == "" {
			name = a.Login
		}
		parts = append(parts, name)
	}
	return strings.Join(parts, " · ")
}

func (i issueItem) FilterValue() string {
	return i.issue.Summary
}

type App struct {
	service     IssueService
	focus       pane
	list        list.Model
	detail      viewport.Model
	issues      []model.Issue
	selected    *model.Issue
	query       string
	err         string
	width       int
	height      int
	ready       bool
	loading     bool
	pageSize    int
	hasMore     bool
	searchInput  textinput.Model
	searching    bool
	issueDialog IssueDialog
	confirmDelete bool
	commenting    bool
	commentInput  textarea.Model
	showHelp      bool
	listCollapsed bool
	listRatio     float64
	settingState  bool
	stateInput    textinput.Model
	assigning     bool
	assignInput   textinput.Model
	finderDialog   FinderDialog
	projectPicker  ProjectPickerDialog
	activeProject  *model.Project
	goingToIssue   bool
	gotoInput      textinput.Model
	gotoProject    string
	restoreIssueID string
	statePath      string
}

func NewApp(service IssueService, state config.State) *App {
	delegate := list.NewDefaultDelegate()
	l := list.New([]list.Item{}, delegate, 0, 0)
	l.Title = ""
	l.SetShowTitle(false)
	l.SetShowHelp(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)

	vp := viewport.New(0, 0)
	vp.SetContent("Loading issues...")

	si := textinput.New()
	si.Placeholder = "YouTrack query (e.g., project: PROJ #Unresolved)"
	si.Prompt = "/ "

	ci := textarea.New()
	ci.Placeholder = "Write a comment..."
	ci.SetHeight(5)
	ci.CharLimit = 5000

	sti := textinput.New()
	sti.Placeholder = "State name (e.g., Open, In Progress, Fixed)"
	sti.Prompt = "State: "

	asi := textinput.New()
	asi.Placeholder = "User login"
	asi.Prompt = "Assign to: "

	gti := textinput.New()
	gti.Placeholder = "Issue number"
	gti.Prompt = "Go to #: "
	gti.Validate = func(s string) error {
		for _, r := range s {
			if r < '0' || r > '9' {
				return fmt.Errorf("digits only")
			}
		}
		return nil
	}

	app := &App{
		service:      service,
		list:         l,
		detail:       vp,
		pageSize:     50,
		listRatio:      state.UI.ListRatio,
		listCollapsed:  state.UI.ListCollapsed,
		restoreIssueID: state.UI.SelectedIssue,
		statePath:      config.DefaultStatePath(),
		searchInput:  si,
		issueDialog: NewIssueDialog(),
		commentInput: ci,
		stateInput:   sti,
		assignInput:  asi,
		finderDialog:  NewFinderDialog(),
		projectPicker: NewProjectPickerDialog(),
		gotoInput:     gti,
	}

	// Restore active project from state
	if state.UI.ActiveProject != "" {
		app.activeProject = &model.Project{ShortName: state.UI.ActiveProject}
	}

	return app
}

func (a *App) Init() tea.Cmd {
	return a.fetchIssuesCmd()
}

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

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.ready = true
		a.resizePanels()
		return a, nil

	case issuesLoadedMsg:
		a.err = ""
		a.loading = false
		a.issues = msg.issues
		a.hasMore = len(msg.issues) == a.pageSize
		items := make([]list.Item, len(msg.issues))
		for i, issue := range msg.issues {
			items[i] = issueItem{issue: issue}
		}
		cmd := a.list.SetItems(items)
		cmds = append(cmds, cmd)
		if len(msg.issues) > 0 {
			targetIdx := 0
			targetID := msg.issues[0].IDReadable
			if a.restoreIssueID != "" {
				for i, issue := range msg.issues {
					if issue.IDReadable == a.restoreIssueID {
						targetIdx = i
						targetID = issue.IDReadable
						break
					}
				}
				a.restoreIssueID = ""
			}
			a.list.Select(targetIdx)
			cmds = append(cmds, a.fetchDetailCmd(targetID))
		} else {
			a.detail.SetContent("No issues found. Press 'c' to create one or '/' to search.")
		}
		return a, tea.Batch(cmds...)

	case moreIssuesLoadedMsg:
		a.loading = false
		a.hasMore = len(msg.issues) == a.pageSize
		a.issues = append(a.issues, msg.issues...)
		items := make([]list.Item, len(a.issues))
		for i, issue := range a.issues {
			items[i] = issueItem{issue: issue}
		}
		cmd := a.list.SetItems(items)
		return a, cmd

	case issueDetailLoadedMsg:
		a.err = ""
		a.loading = false
		a.selected = msg.issue
		a.detail.SetContent(renderIssueDetail(msg.issue))
		a.detail.GotoTop()
		return a, nil

	case projectsLoadedMsg:
		a.loading = false
		cmd := a.issueDialog.OpenCreate(msg.projects)
		if len(msg.projects) > 0 {
			projectID := msg.projects[0].ID
			service := a.service
			return a, tea.Batch(cmd, func() tea.Msg {
				fields, err := service.ListProjectCustomFields(projectID)
				if err != nil {
					return errMsg{err}
				}
				return customFieldsLoadedMsg{fields}
			})
		}
		return a, cmd

	case customFieldsLoadedMsg:
		if a.issueDialog.active {
			currentState := ""
			currentType := ""
			if a.issueDialog.mode == modeEdit && a.selected != nil {
				currentState = a.selected.StateValue()
				currentType = a.selected.TypeValue()
			}
			a.issueDialog.SetCustomFields(msg.fields, currentState, currentType)
		}
		return a, nil

	case assigneeDebounceMsg:
		if a.issueDialog.active && msg.generation == a.issueDialog.assigneeGen {
			query := a.issueDialog.assigneeInput.Value()
			service := a.service
			gen := msg.generation
			return a, func() tea.Msg {
				users, err := service.SearchUsers(query)
				if err != nil {
					return errMsg{err}
				}
				return assigneeSearchResultsMsg{users: users, generation: gen}
			}
		}
		return a, nil

	case assigneeSearchResultsMsg:
		if a.issueDialog.active {
			a.issueDialog.SetAssigneeResults(msg.users, msg.generation)
		}
		return a, nil

	case issueCreatedMsg:
		a.loading = false
		return a, a.fetchIssuesCmd()

	case issueUpdatedMsg:
		a.loading = false
		if a.selected != nil {
			issueID := a.selected.IDReadable
			return a, tea.Batch(a.fetchIssuesCmd(), a.fetchDetailCmd(issueID))
		}
		return a, a.fetchIssuesCmd()

	case issueDeletedMsg:
		a.loading = false
		a.selected = nil
		a.detail.SetContent("Issue deleted.")
		return a, a.fetchIssuesCmd()

	case commentAddedMsg:
		a.loading = false
		if a.selected != nil {
			issueID := a.selected.IDReadable
			return a, a.fetchDetailCmd(issueID)
		}
		return a, nil

	case errMsg:
		a.loading = false
		if a.finderDialog.active {
			a.finderDialog.SetError(msg.err.Error())
			return a, nil
		}
		a.err = msg.err.Error()
		return a, nil

	case finderDebounceMsg:
		if a.finderDialog.active && msg.generation == a.finderDialog.searchGen {
			query := a.finderDialog.Query()
			service := a.service
			gen := msg.generation
			return a, func() tea.Msg {
				issues, err := service.ListIssues(query, 0, 20)
				if err != nil {
					return errMsg{err}
				}
				return finderSearchResultsMsg{issues: issues, generation: gen}
			}
		}
		return a, nil

	case finderSearchResultsMsg:
		if a.finderDialog.active {
			a.finderDialog.SetResults(msg.issues, msg.generation)
		}
		return a, nil

	case projectsForPickerMsg:
		a.loading = false
		if len(msg.projects) == 0 {
			a.err = "No projects found"
			return a, nil
		}
		a.projectPicker.Open(msg.projects)
		return a, nil

	case tea.KeyMsg:
		if m, cmd := a.handleKeyMsg(msg); m != nil {
			return m, cmd
		}
		// Key not handled — fall through to focus-based panel routing below
	}

	// Route updates to focused panel
	var cmd tea.Cmd
	switch a.focus {
	case listPane:
		a.list, cmd = a.list.Update(msg)
		cmds = append(cmds, cmd)

		// Auto-load detail when cursor moves (with loading guard)
		if !a.loading {
			if item, ok := a.list.SelectedItem().(issueItem); ok {
				if a.selected == nil || a.selected.IDReadable != item.issue.IDReadable {
					a.loading = true
					cmds = append(cmds, a.fetchDetailCmd(item.issue.IDReadable))
				}
			}
		}

		// Pagination: load more when near bottom (with loading guard)
		if !a.loading && a.hasMore {
			idx := a.list.Index()
			total := len(a.issues)
			if total > 0 && idx >= total-5 {
				a.loading = true
				cmds = append(cmds, a.fetchMoreIssuesCmd())
			}
		}

	case detailPane:
		a.detail, cmd = a.detail.Update(msg)
		cmds = append(cmds, cmd)
	}

	return a, tea.Batch(cmds...)
}


