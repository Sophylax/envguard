package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const fileName = ".envguard.yml"

// Load finds and parses .envguard.yml from startDir upwards. Missing config returns defaults.
func Load(startDir string) (Config, string, error) {
	cfg := Default()
	path, err := findConfig(startDir)
	if err != nil {
		return Config{}, "", fmt.Errorf("find config: %w", err)
	}
	if path == "" {
		return cfg, "", nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, "", fmt.Errorf("read config %s: %w", path, err)
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, "", fmt.Errorf("parse config %s: %w", path, err)
	}
	return cfg, path, nil
}

func findConfig(startDir string) (string, error) {
	dir, err := filepath.Abs(startDir)
	if err != nil {
		return "", fmt.Errorf("resolve start dir %s: %w", startDir, err)
	}

	for {
		candidate := filepath.Join(dir, fileName)
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		} else if !errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("stat config %s: %w", candidate, err)
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", nil
		}
		dir = parent
	}
}
