package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/cf/lazytrack/internal/api"
	"github.com/cf/lazytrack/internal/config"
)

type setupStep int

const (
	stepURL setupStep = iota
	stepToken
	stepValidating
	stepDone
)

type setupValidateMsg struct {
	err  error
	user string
}

type SetupModel struct {
	urlInput   textinput.Model
	tokenInput textinput.Model
	step       setupStep
	err        string
	cfg        *config.Config
	cancelled  bool
}

func NewSetupModel() *SetupModel {
	urlIn := textinput.New()
	urlIn.Placeholder = "https://youtrack.example.com"
	urlIn.Prompt = "Server URL: "
	urlIn.Focus()

	tokenIn := textinput.New()
	tokenIn.Placeholder = "perm:your-permanent-token"
	tokenIn.Prompt = "Token: "
	tokenIn.EchoMode = textinput.EchoPassword

	return &SetupModel{
		urlInput:   urlIn,
		tokenInput: tokenIn,
		step:       stepURL,
	}
}

func (m *SetupModel) Config() *config.Config {
	return m.cfg
}

func (m *SetupModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m *SetupModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.cancelled = true
			return m, tea.Quit
		case "enter":
			return m.handleEnter()
		case "shift+tab":
			if m.step == stepToken {
				m.step = stepURL
				m.tokenInput.Blur()
				return m, m.urlInput.Focus()
			}
		}

	case setupValidateMsg:
		if msg.err != nil {
			m.step = stepToken
			m.err = msg.err.Error()
			return m, m.tokenInput.Focus()
		}
		// Success — save config
		m.step = stepDone
		serverURL := strings.TrimRight(m.urlInput.Value(), "/")
		token := m.tokenInput.Value()
		m.cfg = &config.Config{
			Server: config.ServerConfig{
				URL:   serverURL,
				Token: token,
			},
		}
		path := config.DefaultPath()
		if err := config.Save(path, m.cfg); err != nil {
			m.err = fmt.Sprintf("Failed to save config: %v", err)
			m.cfg = nil
			return m, nil
		}
		return m, tea.Quit
	}

	var cmd tea.Cmd
	switch m.step {
	case stepURL:
		m.urlInput, cmd = m.urlInput.Update(msg)
	case stepToken:
		m.tokenInput, cmd = m.tokenInput.Update(msg)
	}
	return m, cmd
}

func (m *SetupModel) handleEnter() (tea.Model, tea.Cmd) {
	switch m.step {
	case stepURL:
		if m.urlInput.Value() == "" {
			m.err = "Server URL is required"
			return m, nil
		}
		m.err = ""
		m.step = stepToken
		m.urlInput.Blur()
		return m, m.tokenInput.Focus()

	case stepToken:
		if m.tokenInput.Value() == "" {
			m.err = "Token is required"
			return m, nil
		}
		m.err = ""
		m.step = stepValidating
		// Capture values before closure
		serverURL := strings.TrimRight(m.urlInput.Value(), "/")
		token := m.tokenInput.Value()
		return m, func() tea.Msg {
			client := api.NewClient(serverURL, token)
			user, err := client.GetCurrentUser()
			if err != nil {
				return setupValidateMsg{err: err}
			}
			return setupValidateMsg{user: user.FullName}
		}
	}
	return m, nil
}

func (m *SetupModel) View() string {
	var b strings.Builder

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("99")).
		MarginBottom(1)

	b.WriteString(title.Render("lazytrack — First-Run Setup") + "\n\n")
	b.WriteString("Configure your YouTrack server connection.\n\n")

	b.WriteString(m.urlInput.View() + "\n")

	if m.step >= stepToken {
		b.WriteString(m.tokenInput.View() + "\n")
	}

	b.WriteString("\n")

	if m.step == stepValidating {
		b.WriteString("Validating connection...\n")
	}

	if m.err != "" {
		errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
		b.WriteString(errStyle.Render("Error: "+m.err) + "\n")
	}

	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("enter: next  shift+tab: back  esc: cancel"))

	return b.String()
}
