package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/cf/lazytrack/internal/model"
)

type finderSection int

const (
	finderInputSection finderSection = iota
	finderCheckboxSection
	finderResultsSection
)

// FinderDialog is a modal popup for fuzzy-finding issues by title with type filters.
type FinderDialog struct {
	input          textinput.Model
	filterBug      bool
	filterTask     bool
	checkboxCursor int // 0=Bug, 1=Task
	results        []model.Issue
	resultCursor   int
	focus          finderSection
	active         bool
	submitted      bool
	selectedIssue  *model.Issue
	searchGen      int
	loading        bool
	searchErr      string
}

func NewFinderDialog() FinderDialog {
	ti := textinput.New()
	ti.Placeholder = "Search issues by title..."
	ti.Prompt = "/ "
	ti.CharLimit = 200

	return FinderDialog{
		input:      ti,
		filterBug:  true,
		filterTask: true,
	}
}

func (d *FinderDialog) Open() tea.Cmd {
	d.active = true
	d.submitted = false
	d.selectedIssue = nil
	d.input.SetValue("")
	d.results = nil
	d.resultCursor = 0
	d.focus = finderInputSection
	d.checkboxCursor = 0
	d.searchGen = 0
	d.loading = false
	d.searchErr = ""
	// Note: filter booleans are intentionally NOT reset here so the user's
	// type preferences persist across multiple finder invocations per session.
	return d.input.Focus()
}

func (d *FinderDialog) Close() {
	d.active = false
	d.input.Blur()
}

func (d *FinderDialog) SetResults(issues []model.Issue, gen int) {
	if gen != d.searchGen {
		return // stale results, ignore
	}
	d.results = issues
	d.resultCursor = 0
	d.loading = false
	d.searchErr = ""
}

func (d *FinderDialog) SetError(errStr string) {
	d.loading = false
	d.searchErr = errStr
}

func (d *FinderDialog) Query() string {
	return buildFinderQuery(d.input.Value(), d.filterBug, d.filterTask)
}

func (d *FinderDialog) Update(msg tea.Msg) (FinderDialog, tea.Cmd) {
	if !d.active {
		return *d, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			d.Close()
			return *d, nil

		case "tab":
			d.focus = (d.focus + 1) % 3
			if d.focus == finderInputSection {
				return *d, d.input.Focus()
			}
			d.input.Blur()
			return *d, nil

		case "shift+tab":
			d.focus = (d.focus + 2) % 3
			if d.focus == finderInputSection {
				return *d, d.input.Focus()
			}
			d.input.Blur()
			return *d, nil

		case "enter":
			if d.focus == finderResultsSection && len(d.results) > 0 {
				issue := d.results[d.resultCursor]
				d.selectedIssue = &issue
				d.submitted = true
				d.Close()
				return *d, nil
			}
			// Enter in input section triggers immediate search
			if d.focus == finderInputSection {
				d.searchGen++
				d.loading = true
				d.searchErr = ""
				gen := d.searchGen
				return *d, func() tea.Msg {
					return finderDebounceMsg{generation: gen}
				}
			}
			return *d, nil

		case " ":
			// Only intercept space for checkbox toggling; otherwise let it
			// fall through to the text input handler below.
			if d.focus == finderCheckboxSection {
				switch d.checkboxCursor {
				case 0:
					d.filterBug = !d.filterBug
				case 1:
					d.filterTask = !d.filterTask
				}
				d.searchGen++
				d.loading = true
				d.searchErr = ""
				gen := d.searchGen
				return *d, func() tea.Msg {
					return finderDebounceMsg{generation: gen}
				}
			}
			// Do NOT return here — fall through to text input handler

		case "left", "h":
			if d.focus == finderCheckboxSection {
				if d.checkboxCursor > 0 {
					d.checkboxCursor--
				}
				return *d, nil
			}
			// When focus is on input, let h/left fall through to text input

		case "right", "l":
			if d.focus == finderCheckboxSection {
				if d.checkboxCursor < 1 {
					d.checkboxCursor++
				}
				return *d, nil
			}
			// When focus is on input, let l/right fall through to text input

		case "up", "k":
			if d.focus == finderResultsSection {
				if d.resultCursor > 0 {
					d.resultCursor--
				}
				return *d, nil
			}
			// When focus is on input, let k/up fall through to text input

		case "down", "j":
			if d.focus == finderResultsSection {
				if d.resultCursor < len(d.results)-1 {
					d.resultCursor++
				}
				return *d, nil
			}
			// When focus is on input, let j/down fall through to text input
		}

		// All unhandled key events reach here. Route to text input when focused,
		// and ONLY debounce on actual key events (not blinks/resizes/mouse).
		if d.focus == finderInputSection {
			var cmd tea.Cmd
			d.input, cmd = d.input.Update(msg)

			d.searchGen++
			d.loading = true
			d.searchErr = ""
			gen := d.searchGen
			debounceCmd := tea.Tick(300*time.Millisecond, func(t time.Time) tea.Msg {
				return finderDebounceMsg{generation: gen}
			})
			return *d, tea.Batch(cmd, debounceCmd)
		}

		return *d, nil
	}

	// Non-key messages (blink, resize, mouse): forward to text input for
	// cursor rendering, but do NOT trigger debounce or increment searchGen.
	if d.focus == finderInputSection {
		var cmd tea.Cmd
		d.input, cmd = d.input.Update(msg)
		return *d, cmd
	}

	return *d, nil
}

func (d *FinderDialog) View(width, height int) string {
	if !d.active {
		return ""
	}

	dialogWidth := width * 3 / 5
	if dialogWidth < 50 {
		dialogWidth = 50
	}
	dialogHeight := height * 7 / 10
	if dialogHeight < 15 {
		dialogHeight = 15
	}
	// Cap to terminal height so lipgloss.Place doesn't clip
	if dialogHeight > height-2 {
		dialogHeight = height - 2
	}

	contentWidth := dialogWidth - 6 // padding + border

	// Set text input width to match dialog
	d.input.Width = contentWidth - 3 // account for prompt "/ "

	var b strings.Builder

	// Title
	b.WriteString(titleStyle.Render("Find Issues") + "\n\n")

	// Search input
	b.WriteString(d.input.View() + "\n\n")

	// Type checkboxes
	activeCheckbox := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("69"))

	labels := []struct {
		name    string
		checked bool
	}{
		{"Bug", d.filterBug},
		{"Task", d.filterTask},
	}

	var cbs []string
	for i, l := range labels {
		mark := " "
		if l.checked {
			mark = "x"
		}
		text := fmt.Sprintf("[%s] %s", mark, l.name)
		if d.focus == finderCheckboxSection && d.checkboxCursor == i {
			cbs = append(cbs, activeCheckbox.Render(text))
		} else {
			cbs = append(cbs, text)
		}
	}
	b.WriteString(strings.Join(cbs, "  ") + "\n\n")

	// Results
	resultsHeight := dialogHeight - 10 // space for title, input, checkboxes, hints, padding
	if resultsHeight < 3 {
		resultsHeight = 3
	}

	if d.searchErr != "" {
		b.WriteString(errorStyle.Render("Search error: "+d.searchErr) + "\n")
	} else if d.loading && len(d.results) == 0 {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("Searching...") + "\n")
	} else if len(d.results) == 0 {
		if d.input.Value() != "" {
			b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("No results") + "\n")
		}
	} else {
		normalStyle := lipgloss.NewStyle().Width(contentWidth)
		selectedStyle := lipgloss.NewStyle().
			Width(contentWidth).
			Background(lipgloss.Color("237")).
			Foreground(lipgloss.Color("255"))

		// Scrolling window over results
		start := 0
		if d.resultCursor >= resultsHeight {
			start = d.resultCursor - resultsHeight + 1
		}
		end := start + resultsHeight
		if end > len(d.results) {
			end = len(d.results)
		}

		for i := start; i < end; i++ {
			issue := d.results[i]
			line := fmt.Sprintf("%-12s %s", issue.IDReadable, issue.Summary)
			if len(line) > contentWidth {
				line = line[:contentWidth-1] + "…"
			}
			if d.focus == finderResultsSection && i == d.resultCursor {
				b.WriteString(selectedStyle.Render(line) + "\n")
			} else {
				b.WriteString(normalStyle.Render(line) + "\n")
			}
		}
	}

	// Hint bar
	b.WriteString("\n")
	hint := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	b.WriteString(hint.Render("tab: switch section  enter: select  space: toggle  esc: close"))

	content := b.String()

	dialogStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("99")).
		Padding(1, 2).
		Width(dialogWidth).
		Height(dialogHeight)

	dialog := dialogStyle.Render(content)

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, dialog)
}

// buildFinderQuery constructs a YouTrack query from the finder's search text
// and type filter checkboxes. If all or no types are checked, no Type filter
// is added. Multi-word type values are wrapped in braces per YouTrack syntax.
func buildFinderQuery(text string, bug, task bool) string {
	var parts []string

	allChecked := bug && task
	noneChecked := !bug && !task
	if !allChecked && !noneChecked {
		var types []string
		if bug {
			types = append(types, "Bug")
		}
		if task {
			types = append(types, "Task")
		}
		parts = append(parts, "Type: "+strings.Join(types, ","))
	}

	text = strings.TrimSpace(text)
	if text != "" {
		parts = append(parts, text)
	}

	return strings.Join(parts, " ")
}
