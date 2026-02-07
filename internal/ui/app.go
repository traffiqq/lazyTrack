package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

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
	state := i.issue.StateValue()
	if state == "" {
		return ""
	}
	return stateColor(state)
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
	createDialog CreateDialog
	editDialog   EditDialog
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
	l.Title = "Issues"
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
		createDialog: NewCreateDialog(),
		editDialog:   NewEditDialog(),
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
		a.createDialog.SetProjects(msg.projects)
		cmd := a.createDialog.Open()
		return a, cmd

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
		// When create dialog is active, route input to it
		if a.createDialog.active {
			var cmd tea.Cmd
			a.createDialog, cmd = a.createDialog.Update(msg)
			if a.createDialog.submitted {
				projectID, summary, desc := a.createDialog.Values()
				service := a.service
				a.loading = true
				return a, func() tea.Msg {
					_, err := service.CreateIssue(projectID, summary, desc)
					if err != nil {
						return errMsg{err}
					}
					return issueCreatedMsg{}
				}
			}
			return a, cmd
		}

		// When edit dialog is active, route input to it
		if a.editDialog.active {
			var cmd tea.Cmd
			a.editDialog, cmd = a.editDialog.Update(msg)
			if a.editDialog.submitted {
				summary, desc := a.editDialog.Values()
				issueID := a.editDialog.issueID
				service := a.service
				a.loading = true
				return a, func() tea.Msg {
					err := service.UpdateIssue(issueID, map[string]any{
						"summary":     summary,
						"description": desc,
					})
					if err != nil {
						return errMsg{err}
					}
					return issueUpdatedMsg{}
				}
			}
			return a, cmd
		}

		// When finder is active, route input to it
		if a.finderDialog.active {
			var cmd tea.Cmd
			a.finderDialog, cmd = a.finderDialog.Update(msg)
			if a.finderDialog.submitted && a.finderDialog.selectedIssue != nil {
				issueID := a.finderDialog.selectedIssue.IDReadable
				a.listCollapsed = true
				a.focus = detailPane
				a.resizePanels()
				a.loading = true
				return a, a.fetchDetailCmd(issueID)
			}
			return a, cmd
		}

		// When project picker is active, route input to it
		if a.projectPicker.active {
			var cmd tea.Cmd
			a.projectPicker, cmd = a.projectPicker.Update(msg)
			if a.projectPicker.submitted {
				a.activeProject = a.projectPicker.selectedProject
				a.loading = true
				return a, a.fetchIssuesCmd()
			}
			return a, cmd
		}

		// When going to issue, route input to goto field
		if a.goingToIssue {
			switch msg.String() {
			case "enter":
				num := a.gotoInput.Value()
				if num != "" {
					issueID := a.gotoProject + "-" + num
					a.goingToIssue = false
					a.gotoInput.Blur()
					a.listCollapsed = true
					a.focus = detailPane
					a.resizePanels()
					a.loading = true
					return a, a.fetchDetailCmd(issueID)
				}
				return a, nil
			case "esc":
				a.goingToIssue = false
				a.gotoInput.Blur()
				return a, nil
			default:
				var cmd tea.Cmd
				a.gotoInput, cmd = a.gotoInput.Update(msg)
				return a, cmd
			}
		}

		// Dismiss error on any key
		if a.err != "" {
			a.err = ""
			return a, nil
		}

		// When help is shown, any key dismisses it
		if a.showHelp {
			a.showHelp = false
			return a, nil
		}

		// When confirming delete, handle y/n
		if a.confirmDelete {
			switch msg.String() {
			case "y", "Y":
				a.confirmDelete = false
				if a.selected != nil {
					issueID := a.selected.IDReadable
					service := a.service
					a.loading = true
					return a, func() tea.Msg {
						err := service.DeleteIssue(issueID)
						if err != nil {
							return errMsg{err}
						}
						return issueDeletedMsg{}
					}
				}
			default:
				a.confirmDelete = false
			}
			return a, nil
		}

		// When commenting, route input to comment textarea
		if a.commenting {
			switch msg.String() {
			case "esc":
				a.commenting = false
				a.commentInput.Blur()
				return a, nil
			case "ctrl+d":
				text := a.commentInput.Value()
				if text != "" && a.selected != nil {
					issueID := a.selected.IDReadable
					service := a.service
					a.commenting = false
					a.commentInput.Blur()
					a.commentInput.SetValue("")
					a.loading = true
					return a, func() tea.Msg {
						_, err := service.AddComment(issueID, text)
						if err != nil {
							return errMsg{err}
						}
						return commentAddedMsg{}
					}
				}
				return a, nil
			default:
				var cmd tea.Cmd
				a.commentInput, cmd = a.commentInput.Update(msg)
				return a, cmd
			}
		}

		// When setting state, route input to state input
		if a.settingState {
			switch msg.String() {
			case "enter":
				newState := a.stateInput.Value()
				if newState != "" && a.selected != nil {
					issueID := a.selected.IDReadable
					service := a.service
					a.settingState = false
					a.stateInput.Blur()
					a.loading = true
					return a, func() tea.Msg {
						fields := map[string]any{
							"customFields": []map[string]any{
								{
									"name":  "State",
									"$type": "StateIssueCustomField",
									"value": map[string]string{
										"name":  newState,
										"$type": "StateBundleElement",
									},
								},
							},
						}
						err := service.UpdateIssue(issueID, fields)
						if err != nil {
							return errMsg{err}
						}
						return issueUpdatedMsg{}
					}
				}
				return a, nil
			case "esc":
				a.settingState = false
				a.stateInput.Blur()
				return a, nil
			default:
				var cmd tea.Cmd
				a.stateInput, cmd = a.stateInput.Update(msg)
				return a, cmd
			}
		}

		// When assigning, route input to assign input
		if a.assigning {
			switch msg.String() {
			case "enter":
				login := a.assignInput.Value()
				if login != "" && a.selected != nil {
					issueID := a.selected.IDReadable
					service := a.service
					a.assigning = false
					a.assignInput.Blur()
					a.loading = true
					return a, func() tea.Msg {
						fields := map[string]any{
							"customFields": []map[string]any{
								{
									"name":  "Assignee",
									"$type": "SingleUserIssueCustomField",
									"value": map[string]any{
										"login": login,
										"$type": "User",
									},
								},
							},
						}
						err := service.UpdateIssue(issueID, fields)
						if err != nil {
							return errMsg{err}
						}
						return issueUpdatedMsg{}
					}
				}
				return a, nil
			case "esc":
				a.assigning = false
				a.assignInput.Blur()
				return a, nil
			default:
				var cmd tea.Cmd
				a.assignInput, cmd = a.assignInput.Update(msg)
				return a, cmd
			}
		}

		// When searching, route all input to the search field
		if a.searching {
			switch msg.String() {
			case "enter":
				a.query = a.searchInput.Value()
				a.searching = false
				a.searchInput.Blur()
				a.loading = true
				return a, a.fetchIssuesCmd()
			case "esc":
				a.searching = false
				a.searchInput.Blur()
				return a, nil
			default:
				var cmd tea.Cmd
				a.searchInput, cmd = a.searchInput.Update(msg)
				return a, cmd
			}
		}

		switch msg.String() {
		case "ctrl+c", "q":
			state := config.State{
				UI: config.UIState{
					ListRatio:     a.listRatio,
					ListCollapsed: a.listCollapsed,
				},
			}
			if a.selected != nil {
				state.UI.SelectedIssue = a.selected.IDReadable
			}
			if a.activeProject != nil {
				state.UI.ActiveProject = a.activeProject.ShortName
			}
			_ = config.SaveStateToPath(a.statePath, state)
			return a, tea.Quit
		case "tab":
			if a.focus == listPane {
				a.focus = detailPane
			} else {
				a.focus = listPane
			}
			return a, nil
		case "enter":
			if a.focus == listPane {
				if item, ok := a.list.SelectedItem().(issueItem); ok {
					a.focus = detailPane
					a.loading = true
					return a, a.fetchDetailCmd(item.issue.IDReadable)
				}
			}
		case "/":
			a.searching = true
			a.searchInput.SetValue(a.query)
			return a, a.searchInput.Focus()
		case "c":
			a.loading = true
			service := a.service
			return a, func() tea.Msg {
				projects, err := service.ListProjects()
				if err != nil {
					return errMsg{err}
				}
				return projectsLoadedMsg{projects}
			}
		case "e":
			if a.selected != nil {
				cmd := a.editDialog.Open(a.selected.IDReadable, a.selected.Summary, a.selected.Description)
				return a, cmd
			}
		case "d":
			if a.selected != nil {
				a.confirmDelete = true
				return a, nil
			}
		case "C":
			if a.selected != nil {
				a.commenting = true
				a.commentInput.SetValue("")
				return a, a.commentInput.Focus()
			}
		case "ctrl+e":
			a.listCollapsed = !a.listCollapsed
			if a.listCollapsed {
				a.focus = detailPane
			} else {
				a.focus = listPane
			}
			a.resizePanels()
			return a, nil
		case "ctrl+right", "L":
			if !a.listCollapsed {
				a.listRatio += 0.02
				if a.listRatio > 0.8 {
					a.listRatio = 0.8
				}
				a.resizePanels()
			}
			return a, nil
		case "ctrl+left", "H":
			if !a.listCollapsed {
				a.listRatio -= 0.02
				if a.listRatio < 0.2 {
					a.listRatio = 0.2
				}
				a.resizePanels()
			}
			return a, nil
		case "?":
			a.showHelp = !a.showHelp
			return a, nil
		case "s":
			if a.selected != nil {
				a.settingState = true
				a.stateInput.SetValue("")
				return a, a.stateInput.Focus()
			}
		case "a":
			if a.selected != nil {
				a.assigning = true
				a.assignInput.SetValue("")
				return a, a.assignInput.Focus()
			}
		case "#":
			proj := a.resolveGotoProject()
			if proj == "" {
				a.err = "Select a project first (press p)"
				return a, nil
			}
			a.gotoProject = proj
			a.goingToIssue = true
			a.gotoInput.SetValue("")
			a.gotoInput.Prompt = fmt.Sprintf("Go to %s-#: ", proj)
			return a, a.gotoInput.Focus()
		case "p":
			a.loading = true
			service := a.service
			return a, func() tea.Msg {
				projects, err := service.ListProjects()
				if err != nil {
					return errMsg{err}
				}
				return projectsForPickerMsg{projects}
			}
		case "f":
			return a, a.finderDialog.Open()
		}
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

func (a *App) View() string {
	if !a.ready {
		return "Loading..."
	}

	// Render overlays if active
	if a.finderDialog.active {
		return a.finderDialog.View(a.width, a.height)
	}
	if a.projectPicker.active {
		return a.projectPicker.View(a.width, a.height)
	}
	if a.showHelp {
		return renderHelp(a.width, a.height)
	}
	if a.createDialog.active {
		return a.createDialog.View(a.width, a.height)
	}
	if a.editDialog.active {
		return a.editDialog.View(a.width, a.height)
	}

	var panels string
	panelHeight := a.height - 3

	if a.listCollapsed {
		innerWidth := a.width - 2
		rightPanel := focusedPanelStyle.
			Width(innerWidth).
			Height(panelHeight).
			Render(a.detail.View())
		panels = rightPanel
	} else {
		listWidth := int(float64(a.width) * a.listRatio)
		detailWidth := a.width - listWidth

		leftStyle := panelStyle
		rightStyle := panelStyle
		if a.focus == listPane {
			leftStyle = focusedPanelStyle
		} else {
			rightStyle = focusedPanelStyle
		}

		innerListWidth := listWidth - 2
		innerDetailWidth := detailWidth - 2

		leftPanel := leftStyle.
			Width(innerListWidth).
			Height(panelHeight).
			Render(a.list.View())

		rightPanel := rightStyle.
			Width(innerDetailWidth).
			Height(panelHeight).
			Render(a.detail.View())

		panels = lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel)

		// Comment overlay replaces right panel
		if a.commenting {
			commentView := lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("99")).
				Padding(1, 2).
				Width(detailWidth - 2).
				Render(
					titleStyle.Render("Add Comment")+"\n\n"+
						a.commentInput.View()+"\n\n"+
						lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("ctrl+d: submit  esc: cancel"),
				)
			panels = lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, commentView)
		}
	}

	var bottom string
	if a.searching {
		bottom = a.searchInput.View()
	} else if a.settingState {
		bottom = a.stateInput.View()
	} else if a.assigning {
		bottom = a.assignInput.View()
	} else if a.goingToIssue {
		bottom = a.gotoInput.View()
	} else if a.confirmDelete && a.selected != nil {
		bottom = errorStyle.Render(fmt.Sprintf("Delete %s? (y/n)", a.selected.IDReadable))
	} else {
		bottom = a.renderStatusBar()
	}

	return lipgloss.JoinVertical(lipgloss.Left, panels, bottom)
}

func (a *App) resizePanels() {
	panelHeight := a.height - 5

	if a.listCollapsed {
		a.detail.Width = a.width - 4
		a.detail.Height = panelHeight
	} else {
		listOuter := int(float64(a.width) * a.listRatio)
		listWidth := listOuter - 4
		detailWidth := a.width - listOuter - 4
		a.list.SetSize(listWidth, panelHeight)
		a.detail.Width = detailWidth
		a.detail.Height = panelHeight
	}
}

func (a *App) renderStatusBar() string {
	if a.err != "" {
		return lipgloss.NewStyle().
			Background(lipgloss.Color("236")).
			Padding(0, 1).
			Width(a.width - 2).
			Render(errorStyle.Render("Error: " + a.err))
	}

	// Left side: app name + context
	left := titleStyle.Render(iconApp + " lazytrack")
	if a.activeProject != nil {
		left += hintDescStyle.Render(" | project: " + a.activeProject.ShortName)
	}
	if a.query != "" {
		left += hintDescStyle.Render(" | query: " + a.query)
	}
	if a.loading {
		left += keyStyle.Render(" | loading...")
	}

	// Right side: mode-aware hints
	hints := modeHints(a.commenting, a.focus)
	rightParts := make([]string, len(hints))
	for i, h := range hints {
		rightParts[i] = formatKeyHint(h.key, h.desc)
	}

	// Overflow: drop hints from right until content fits within available width
	leftWidth := lipgloss.Width(left)
	availWidth := a.width - 2 // content width inside statusBarStyle padding (0,1)
	for len(rightParts) > 0 {
		right := strings.Join(rightParts, "  ")
		if leftWidth+2+lipgloss.Width(right) <= availWidth {
			break
		}
		rightParts = rightParts[:len(rightParts)-1]
	}

	right := strings.Join(rightParts, "  ")
	rightWidth := lipgloss.Width(right)
	gap := availWidth - leftWidth - rightWidth
	if gap < 0 {
		gap = 0
	}

	content := left + strings.Repeat(" ", gap) + right
	return statusBarStyle.Width(availWidth).Render(content)
}
