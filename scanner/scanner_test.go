package scanner

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/sophylax/envguard/allowlist"
	"github.com/sophylax/envguard/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScanDirtyFixture(t *testing.T) {
	cfg := config.Default()
	cfg.ExcludePaths = nil

	engine, err := NewEngine(cfg, allowlist.Set{})
	require.NoError(t, err)

	findings, err := engine.ScanPaths([]string{filepath.Join("..", "testdata", "dirty.js")})
	require.NoError(t, err)

	assert.Len(t, findings, 6)
	require.NotEmpty(t, findings)
	assert.Equal(t, "HIGH", findings[0].Severity)
	assert.Equal(t, "pattern", findings[0].Source)

	severities := make(map[string]int)
	rules := make(map[string]int)
	for _, finding := range findings {
		severities[finding.Severity]++
		rules[finding.RuleName]++
		assert.NotEmpty(t, finding.Fingerprint)
	}

	assert.Equal(t, 4, severities["HIGH"])
	assert.Equal(t, 2, severities["MEDIUM"])
	assert.Equal(t, 1, rules["AWS Access Key"])
	assert.Equal(t, 1, rules["Slack Token"])
	assert.Equal(t, 1, rules["Generic API Key"])
	assert.Equal(t, 1, rules["Database URL"])
}

func TestScanWarnsWhenFileExceedsMaxSize(t *testing.T) {
	tempDir := t.TempDir()
	target := filepath.Join(tempDir, "large.txt")
	require.NoError(t, os.WriteFile(target, bytes.Repeat([]byte("a"), 2048), 0o644))

	cfg := config.Default()
	cfg.MaxFileSizeKB = 1

	engine, err := NewEngine(cfg, allowlist.Set{})
	require.NoError(t, err)

	findings, err := engine.ScanPaths([]string{target})
	require.NoError(t, err)
	assert.Empty(t, findings)
	require.Len(t, engine.Warnings(), 1)
	assert.Contains(t, engine.Warnings()[0], "exceeds limit")
}

func TestScanSkipsMissingPathWithWarning(t *testing.T) {
	tempDir := t.TempDir()
	target := filepath.Join(tempDir, "deleted.txt")

	cfg := config.Default()
	engine, err := NewEngine(cfg, allowlist.Set{})
	require.NoError(t, err)

	findings, err := engine.ScanPaths([]string{target})
	require.NoError(t, err)
	assert.Empty(t, findings)
	require.Len(t, engine.Warnings(), 1)
	assert.Contains(t, engine.Warnings()[0], "path no longer exists")
	assert.Contains(t, engine.Warnings()[0], "deleted.txt")
}

func TestEnvFileDetectionCanBeAllowlisted(t *testing.T) {
	tempDir := t.TempDir()
	chdirForTest(t, tempDir)

	target := filepath.Join(tempDir, ".env")
	require.NoError(t, os.WriteFile(target, []byte("HELLO=world\n"), 0o644))

	cfg := config.Default()
	cfg.ExcludePaths = nil

	engine, err := NewEngine(cfg, allowlist.Set{})
	require.NoError(t, err)

	findings, err := engine.ScanPaths([]string{target})
	require.NoError(t, err)
	require.Len(t, findings, 1)
	assert.Equal(t, ".env file staged", findings[0].RuleName)

	allow := allowlist.Set{findings[0].Fingerprint: struct{}{}}
	engine, err = NewEngine(cfg, allow)
	require.NoError(t, err)

	findings, err = engine.ScanPaths([]string{target})
	require.NoError(t, err)
	assert.Empty(t, findings)
}

func chdirForTest(t *testing.T, dir string) {
	t.Helper()
	wd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() {
		require.NoError(t, os.Chdir(wd))
	})
}
