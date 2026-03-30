package scanner

import (
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
