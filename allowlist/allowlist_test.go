package allowlist

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddAndLoadAllowlist(t *testing.T) {
	repoRoot := t.TempDir()

	path, err := Add(repoRoot, "abc123")
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(repoRoot, ".envguard-ignore"), path)

	set, _, err := Load(repoRoot)
	require.NoError(t, err)
	assert.True(t, set.Contains("abc123"))

	_, err = Add(repoRoot, "abc123")
	require.NoError(t, err)

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "abc123\n", string(data))
}
