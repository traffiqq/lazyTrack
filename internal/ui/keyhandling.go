package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/cf/lazytrack/internal/config"
)

// handleKeyMsg routes all tea.KeyMsg events. Called from Update.
// Returns (nil, nil) when the key is not handled, signaling Update
// to fall through to focus-based panel routing (list/viewport navigation).
func (a *App) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// When issue dialog is active, route input to it
	if a.issueDialog.active {
		var cmd tea.Cmd
		a.issueDialog, cmd = a.issueDialog.Update(msg)
		if a.issueDialog.submitted {
			service := a.service
			a.loading = true
			customFields := a.issueDialog.buildCustomFields()
			if a.issueDialog.mode == modeCreate {
				projectID := ""
				if len(a.issueDialog.projects) > 0 {
					projectID = a.issueDialog.projects[a.issueDialog.projectIndex].ID
				}
				summary := a.issueDialog.summaryInput.Value()
				desc := a.issueDialog.descInput.Value()
				return a, func() tea.Msg {
					_, err := service.CreateIssue(projectID, summary, desc, customFields)
					if err != nil {
						return errMsg{err}
					}
					return issueCreatedMsg{}
				}
			} else {
				issueID := a.issueDialog.issueID
				summary := a.issueDialog.summaryInput.Value()
				desc := a.issueDialog.descInput.Value()
				return a, func() tea.Msg {
					fields := map[string]any{
						"summary":      summary,
						"description":  desc,
						"customFields": customFields,
					}
					err := service.UpdateIssue(issueID, fields)
					if err != nil {
						return errMsg{err}
					}
					return issueUpdatedMsg{}
				}
			}
		}
		if a.issueDialog.projectChanged {
			a.issueDialog.projectChanged = false
			if len(a.issueDialog.projects) > 0 {
				projectID := a.issueDialog.projects[a.issueDialog.projectIndex].ID
				service := a.service
				return a, tea.Batch(cmd, func() tea.Msg {
					fields, err := service.ListProjectCustomFields(projectID)
					if err != nil {
						return errMsg{err}
					}
					return customFieldsLoadedMsg{fields}
				})
			}
		}
		return a, cmd
	}

	// When notification dialog is active, route input to it
	if a.notifDialog.active {
		var cmd tea.Cmd
		a.notifDialog, cmd = a.notifDialog.Update(msg)
		if a.notifDialog.submitted && a.notifDialog.selectedIssue != nil {
			issueID := a.notifDialog.selectedIssue.IDReadable
			a.lastCheckedMentions = latestIssueTimestamp(a.mentionedIssues)
			a.unreadMentionCount = 0
			a.listCollapsed = true
			a.focus = detailPane
			a.resizePanels()
			a.loading = true
			return a, a.fetchDetailCmd(issueID)
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
		case "ctrl+s":
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
				ListRatio:           a.listRatio,
				ListCollapsed:       a.listCollapsed,
				LastCheckedMentions: a.lastCheckedMentions,
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
		hasComments := a.selected != nil && len(a.selected.Comments) > 0
		switch a.focus {
		case listPane:
			a.focus = detailPane
		case detailPane:
			if hasComments {
				a.focus = commentsPane
			} else {
				a.focus = listPane
			}
		case commentsPane:
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
			issue := a.selected
			comments := issue.Comments
			cmd := a.issueDialog.OpenEdit(issue, comments)
			if issue.Project != nil {
				projectID := issue.Project.ID
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
	case "n":
		if a.currentUser == nil {
			a.err = "Could not load user — mentions unavailable"
			return a, nil
		}
		a.notifDialog.Open(a.lastCheckedMentions)
		service := a.service
		query := "mentioned: me sort by: updated desc"
		if a.activeProject != nil {
			query = "project: " + a.activeProject.ShortName + " " + query
		}
		return a, func() tea.Msg {
			issues, err := service.ListIssues(query, 0, 50)
			if err != nil {
				return errMsg{err}
			}
			return mentionsLoadedMsg{issues}
		}
	case "r":
		a.loading = true
		var refreshCmds []tea.Cmd
		refreshCmds = append(refreshCmds, a.fetchIssuesCmd())
		if a.selected != nil {
			issueID := a.selected.IDReadable
			refreshCmds = append(refreshCmds, a.fetchDetailCmd(issueID))
		}
		if a.currentUser != nil {
			refreshCmds = append(refreshCmds, a.fetchMentionsCmd())
		}
		return a, tea.Batch(refreshCmds...)
	}

	// Key not handled — return nil to signal Update to fall through
	// to focus-based panel routing (j/k, arrows, pgup/pgdn, etc.)
	return nil, nil
}
