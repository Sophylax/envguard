package git

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInstallHookNonInteractiveForeignHookRequiresYes(t *testing.T) {
	repoRoot := initTestRepo(t)
	hookPath := HookPath(repoRoot)
	require.NoError(t, os.WriteFile(hookPath, []byte("#!/bin/sh\necho foreign\n"), 0o755))

	var output bytes.Buffer
	_, err := InstallHook(repoRoot, strings.NewReader(""), &output, InstallOptions{Interactive: false})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "rerun with --yes")
}

func TestInstallHookForceMergesForeignHook(t *testing.T) {
	repoRoot := initTestRepo(t)
	hookPath := HookPath(repoRoot)
	require.NoError(t, os.WriteFile(hookPath, []byte("#!/bin/sh\necho foreign\n"), 0o755))

	var output bytes.Buffer
	installedPath, err := InstallHook(repoRoot, strings.NewReader(""), &output, InstallOptions{Force: true, Interactive: false})
	require.NoError(t, err)
	assert.Equal(t, hookPath, installedPath)

	data, err := os.ReadFile(hookPath)
	require.NoError(t, err)
	content := string(data)
	assert.Contains(t, content, "envguard check")
	assert.Contains(t, content, "echo foreign")
}

func TestInstallHookInteractiveDeclineKeepsForeignHook(t *testing.T) {
	repoRoot := initTestRepo(t)
	hookPath := HookPath(repoRoot)
	original := "#!/bin/sh\necho foreign\n"
	require.NoError(t, os.WriteFile(hookPath, []byte(original), 0o755))

	var output bytes.Buffer
	_, err := InstallHook(repoRoot, strings.NewReader("n\n"), &output, InstallOptions{Interactive: true})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "installation cancelled")

	data, err := os.ReadFile(hookPath)
	require.NoError(t, err)
	assert.Equal(t, original, string(data))
}

func TestInstallHookOverwritesExistingEnvguardBlockWithoutDuplication(t *testing.T) {
	repoRoot := initTestRepo(t)
	hookPath := HookPath(repoRoot)
	existing := "#!/bin/sh\n" + hookScript + "\n" + "echo foreign\n"
	require.NoError(t, os.WriteFile(hookPath, []byte(existing), 0o755))

	var output bytes.Buffer
	installedPath, err := InstallHook(repoRoot, strings.NewReader(""), &output, InstallOptions{})
	require.NoError(t, err)
	assert.Equal(t, hookPath, installedPath)

	data, err := os.ReadFile(hookPath)
	require.NoError(t, err)
	content := string(data)
	assert.Equal(t, 1, strings.Count(content, hookMarker))
	assert.Contains(t, content, "envguard check")
	assert.Contains(t, content, "echo foreign")
}

func initTestRepo(t *testing.T) string {
	t.Helper()
	repoRoot := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(repoRoot, ".git", "hooks"), 0o755))
	return repoRoot
}
