package config

import (
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
	"gopkg.in/yaml.v3"
)

type State struct {
	UI UIState `yaml:"ui"`
}

type UIState struct {
	ListRatio     float64 `yaml:"list_ratio"`
	ListCollapsed bool    `yaml:"list_collapsed"`
	SelectedIssue string  `yaml:"selected_issue"`
}

// DefaultState returns a State with sensible defaults.
func DefaultState() State {
	return State{
		UI: UIState{
			ListRatio: 0.4,
		},
	}
}

// DefaultStatePath returns the XDG-compliant state file path.
func DefaultStatePath() string {
	return filepath.Join(xdg.StateHome, "lazytrack", "state.yaml")
}

// LoadStateFromPath reads and parses the state file at the given path.
// Returns default state if the file is missing or invalid.
func LoadStateFromPath(path string) State {
	data, err := os.ReadFile(path)
	if err != nil {
		return DefaultState()
	}

	var state State
	if err := yaml.Unmarshal(data, &state); err != nil {
		return DefaultState()
	}

	if state.UI.ListRatio < 0.2 || state.UI.ListRatio > 0.8 {
		state.UI.ListRatio = 0.4
	}

	return state
}

// LoadState reads the state from the default XDG path.
func LoadState() State {
	return LoadStateFromPath(DefaultStatePath())
}
