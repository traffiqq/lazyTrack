package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/cf/lazytrack/internal/api"
	"github.com/cf/lazytrack/internal/config"
	"github.com/cf/lazytrack/internal/ui"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		setup := ui.NewSetupModel()
		p := tea.NewProgram(setup)
		result, err := p.Run()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Setup error: %v\n", err)
			os.Exit(1)
		}
		setupModel, ok := result.(*ui.SetupModel)
		if !ok || setupModel.Config() == nil {
			fmt.Fprintf(os.Stderr, "Setup cancelled.\n")
			os.Exit(0)
		}
		cfg = setupModel.Config()
	}

	client := api.NewClient(cfg.Server.URL, cfg.Server.Token)

	app := ui.NewApp(client)
	p := tea.NewProgram(app, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
