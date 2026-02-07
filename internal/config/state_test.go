package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadState_Valid(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "state.yaml")

	content := []byte(`ui:
  list_ratio: 0.35
  list_collapsed: true
  selected_issue: "PROJ-123"
`)
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatal(err)
	}

	state := LoadStateFromPath(path)

	if state.UI.ListRatio != 0.35 {
		t.Errorf("got ListRatio %v, want 0.35", state.UI.ListRatio)
	}
	if !state.UI.ListCollapsed {
		t.Error("got ListCollapsed false, want true")
	}
	if state.UI.SelectedIssue != "PROJ-123" {
		t.Errorf("got SelectedIssue %q, want %q", state.UI.SelectedIssue, "PROJ-123")
	}
}

func TestLoadState_MissingFile(t *testing.T) {
	state := LoadStateFromPath("/nonexistent/state.yaml")

	if state.UI.ListRatio != 0.4 {
		t.Errorf("got ListRatio %v, want default 0.4", state.UI.ListRatio)
	}
	if state.UI.ListCollapsed {
		t.Error("got ListCollapsed true, want default false")
	}
	if state.UI.SelectedIssue != "" {
		t.Errorf("got SelectedIssue %q, want empty", state.UI.SelectedIssue)
	}
}

func TestLoadState_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "state.yaml")

	if err := os.WriteFile(path, []byte(":::invalid"), 0644); err != nil {
		t.Fatal(err)
	}

	state := LoadStateFromPath(path)

	if state.UI.ListRatio != 0.4 {
		t.Errorf("got ListRatio %v, want default 0.4", state.UI.ListRatio)
	}
}

func TestLoadState_RatioOutOfBounds(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "state.yaml")

	content := []byte(`ui:
  list_ratio: 0.95
`)
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatal(err)
	}

	state := LoadStateFromPath(path)

	if state.UI.ListRatio != 0.4 {
		t.Errorf("got ListRatio %v, want clamped default 0.4", state.UI.ListRatio)
	}
}

func TestDefaultStatePath(t *testing.T) {
	path := DefaultStatePath()
	if path == "" {
		t.Fatal("expected non-empty default state path")
	}
}

func TestSaveState_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "state.yaml")

	state := State{
		UI: UIState{
			ListRatio:     0.6,
			ListCollapsed: true,
			SelectedIssue: "ABC-99",
		},
	}

	if err := SaveStateToPath(path, state); err != nil {
		t.Fatalf("save error: %v", err)
	}

	loaded := LoadStateFromPath(path)

	if loaded.UI.ListRatio != 0.6 {
		t.Errorf("got ListRatio %v, want 0.6", loaded.UI.ListRatio)
	}
	if !loaded.UI.ListCollapsed {
		t.Error("got ListCollapsed false, want true")
	}
	if loaded.UI.SelectedIssue != "ABC-99" {
		t.Errorf("got SelectedIssue %q, want %q", loaded.UI.SelectedIssue, "ABC-99")
	}
}

func TestLoadState_ActiveProject(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "state.yaml")

	content := []byte(`ui:
  list_ratio: 0.4
  active_project: "PROJ"
`)
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatal(err)
	}

	state := LoadStateFromPath(path)

	if state.UI.ActiveProject != "PROJ" {
		t.Errorf("got ActiveProject %q, want %q", state.UI.ActiveProject, "PROJ")
	}
}

func TestSaveState_ActiveProject_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "state.yaml")

	state := State{
		UI: UIState{
			ListRatio:     0.4,
			ActiveProject: "ABC",
		},
	}

	if err := SaveStateToPath(path, state); err != nil {
		t.Fatalf("save error: %v", err)
	}

	loaded := LoadStateFromPath(path)

	if loaded.UI.ActiveProject != "ABC" {
		t.Errorf("got ActiveProject %q, want %q", loaded.UI.ActiveProject, "ABC")
	}
}

func TestSaveState_CreatesDirectory(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nested", "dir", "state.yaml")

	state := DefaultState()
	if err := SaveStateToPath(path, state); err != nil {
		t.Fatalf("save error: %v", err)
	}

	loaded := LoadStateFromPath(path)
	if loaded.UI.ListRatio != 0.4 {
		t.Errorf("got ListRatio %v, want 0.4", loaded.UI.ListRatio)
	}
}
