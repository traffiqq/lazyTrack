package config

import (
	"fmt"
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
	ActiveProject       string  `yaml:"active_project,omitempty"`
	LastCheckedMentions int64   `yaml:"last_checked_mentions,omitempty"`
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

// SaveStateToPath writes the state to the given path, creating directories as needed.
func SaveStateToPath(path string, state State) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating state directory: %w", err)
	}

	data, err := yaml.Marshal(state)
	if err != nil {
		return fmt.Errorf("marshaling state: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("writing state: %w", err)
	}

	return nil
}

// SaveState writes the state to the default XDG path.
func SaveState(state State) error {
	return SaveStateToPath(DefaultStatePath(), state)
}
