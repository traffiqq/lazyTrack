package config

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime"

	"github.com/adrg/xdg"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Server ServerConfig `yaml:"server"`
}

type ServerConfig struct {
	URL   string `yaml:"url"`
	Token string `yaml:"token"`
}

// Validate checks that required fields are present and valid.
func (c *Config) Validate() error {
	if c.Server.URL == "" {
		return fmt.Errorf("server.url is required")
	}
	if _, err := url.Parse(c.Server.URL); err != nil {
		return fmt.Errorf("server.url is invalid: %w", err)
	}
	if c.Server.Token == "" {
		return fmt.Errorf("server.token is required")
	}
	return nil
}

// DefaultPath returns the XDG-compliant config file path.
func DefaultPath() string {
	return filepath.Join(xdg.ConfigHome, "lazytrack", "config.yaml")
}

// LoadFromPath reads and parses the config file at the given path.
func LoadFromPath(path string) (*Config, error) {
	// Check file permissions on non-Windows platforms
	if runtime.GOOS != "windows" {
		info, err := os.Stat(path)
		if err != nil {
			return nil, fmt.Errorf("reading config: %w", err)
		}
		if info.Mode().Perm()&0077 != 0 {
			return nil, fmt.Errorf("config file %s has too-open permissions %v, expected 0600 — run: chmod 600 %s", path, info.Mode().Perm(), path)
		}
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &cfg, nil
}

// Load reads the config from the default XDG path.
func Load() (*Config, error) {
	return LoadFromPath(DefaultPath())
}

// Save writes the config to the given path, creating directories as needed.
func Save(path string, cfg *Config) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}

	return nil
}
