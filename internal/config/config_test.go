package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	content := []byte(`server:
  url: "https://youtrack.example.com"
  token: "perm:test-token"
`)
	if err := os.WriteFile(path, content, 0600); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadFromPath(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Server.URL != "https://youtrack.example.com" {
		t.Errorf("got URL %q, want %q", cfg.Server.URL, "https://youtrack.example.com")
	}
	if cfg.Server.Token != "perm:test-token" {
		t.Errorf("got Token %q, want %q", cfg.Server.Token, "perm:test-token")
	}
}

func TestLoadConfig_FileNotFound(t *testing.T) {
	_, err := LoadFromPath("/nonexistent/config.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoadConfig_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	if err := os.WriteFile(path, []byte(":::invalid"), 0600); err != nil {
		t.Fatal(err)
	}

	_, err := LoadFromPath(path)
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestValidate_MissingURL(t *testing.T) {
	cfg := &Config{Server: ServerConfig{Token: "perm:test"}}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for missing URL")
	}
}

func TestValidate_MissingToken(t *testing.T) {
	cfg := &Config{Server: ServerConfig{URL: "https://example.com"}}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for missing token")
	}
}

func TestValidate_Valid(t *testing.T) {
	cfg := &Config{Server: ServerConfig{URL: "https://example.com", Token: "perm:test"}}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSave_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "subdir", "config.yaml")

	cfg := &Config{Server: ServerConfig{URL: "https://example.com", Token: "perm:abc"}}
	if err := Save(path, cfg); err != nil {
		t.Fatalf("save error: %v", err)
	}

	loaded, err := LoadFromPath(path)
	if err != nil {
		t.Fatalf("load error: %v", err)
	}
	if loaded.Server.URL != cfg.Server.URL {
		t.Errorf("got URL %q, want %q", loaded.Server.URL, cfg.Server.URL)
	}
	if loaded.Server.Token != cfg.Server.Token {
		t.Errorf("got Token %q, want %q", loaded.Server.Token, cfg.Server.Token)
	}

	// Verify file permissions are restricted
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat error: %v", err)
	}
	if info.Mode().Perm() != 0600 {
		t.Errorf("got permissions %v, want 0600", info.Mode().Perm())
	}
}

func TestDefaultPath(t *testing.T) {
	path := DefaultPath()
	if path == "" {
		t.Fatal("expected non-empty default path")
	}
}
