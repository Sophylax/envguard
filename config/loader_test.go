package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadFindsConfigFromParentDirectory(t *testing.T) {
	repoRoot := t.TempDir()
	nestedDir := filepath.Join(repoRoot, "a", "b", "c")
	require.NoError(t, os.MkdirAll(nestedDir, 0o755))

	configPath := filepath.Join(repoRoot, ".envguard.yml")
	configBody := []byte("entropy_threshold: 5.1\nmin_length: 12\nallow_test_fixtures: true\n")
	require.NoError(t, os.WriteFile(configPath, configBody, 0o644))

	cfg, loadedPath, err := Load(nestedDir)
	require.NoError(t, err)

	assert.Equal(t, configPath, loadedPath)
	assert.Equal(t, 5.1, cfg.EntropyThreshold)
	assert.Equal(t, 12, cfg.MinLength)
	assert.True(t, cfg.AllowTestFixtures)
}
