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
	hookPath, err := HookPath(repoRoot)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(hookPath, []byte("#!/bin/sh\necho foreign\n"), 0o755))

	var output bytes.Buffer
	_, err = InstallHook(repoRoot, strings.NewReader(""), &output, InstallOptions{BinaryPath: "/usr/local/bin/envguard", Interactive: false})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "rerun with --yes")
}

func TestInstallHookForceMergesForeignHook(t *testing.T) {
	repoRoot := initTestRepo(t)
	hookPath, err := HookPath(repoRoot)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(hookPath, []byte("#!/bin/sh\necho foreign\n"), 0o755))

	var output bytes.Buffer
	installedPath, err := InstallHook(repoRoot, strings.NewReader(""), &output, InstallOptions{BinaryPath: "/usr/local/bin/envguard", Force: true, Interactive: false})
	require.NoError(t, err)
	assert.Equal(t, hookPath, installedPath)

	data, err := os.ReadFile(hookPath)
	require.NoError(t, err)
	content := string(data)
	assert.Contains(t, content, "ENVGUARD_BIN='/usr/local/bin/envguard'")
	assert.Contains(t, content, "if command -v envguard >/dev/null 2>&1; then")
	assert.Contains(t, content, "elif [ -n \"$ENVGUARD_BIN\" ] && [ -x \"$ENVGUARD_BIN\" ]; then")
	assert.Less(t, strings.Index(content, "envguard check"), strings.Index(content, `"$ENVGUARD_BIN" check`))
	assert.Contains(t, content, "echo foreign")
}

func TestInstallHookInteractiveDeclineKeepsForeignHook(t *testing.T) {
	repoRoot := initTestRepo(t)
	hookPath, err := HookPath(repoRoot)
	require.NoError(t, err)
	original := "#!/bin/sh\necho foreign\n"
	require.NoError(t, os.WriteFile(hookPath, []byte(original), 0o755))

	var output bytes.Buffer
	_, err = InstallHook(repoRoot, strings.NewReader("n\n"), &output, InstallOptions{BinaryPath: "/usr/local/bin/envguard", Interactive: true})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "installation cancelled")

	data, err := os.ReadFile(hookPath)
	require.NoError(t, err)
	assert.Equal(t, original, string(data))
}

func TestInstallHookOverwritesExistingEnvguardBlockWithoutDuplication(t *testing.T) {
	repoRoot := initTestRepo(t)
	hookPath, err := HookPath(repoRoot)
	require.NoError(t, err)
	existing := "#!/bin/sh\n" + buildHookScript("/old/bin/envguard") + "\n" + "echo foreign\n"
	require.NoError(t, os.WriteFile(hookPath, []byte(existing), 0o755))

	var output bytes.Buffer
	installedPath, err := InstallHook(repoRoot, strings.NewReader(""), &output, InstallOptions{BinaryPath: "/new/bin/envguard"})
	require.NoError(t, err)
	assert.Equal(t, hookPath, installedPath)

	data, err := os.ReadFile(hookPath)
	require.NoError(t, err)
	content := string(data)
	assert.Equal(t, 1, strings.Count(content, hookMarker))
	assert.Equal(t, 1, strings.Count(content, hookMarkerEnd))
	assert.Contains(t, content, "ENVGUARD_BIN='/new/bin/envguard'")
	assert.NotContains(t, content, "ENVGUARD_BIN='/old/bin/envguard'")
	assert.Less(t, strings.Index(content, "envguard check"), strings.Index(content, `"$ENVGUARD_BIN" check`))
	assert.Contains(t, content, "echo foreign")
}

func TestUninstallHookRemovesNewStyleEnvguardBlock(t *testing.T) {
	repoRoot := initTestRepo(t)
	hookPath, err := HookPath(repoRoot)
	require.NoError(t, err)
	existing := "#!/bin/sh\n" + buildHookScript("/usr/local/bin/envguard") + "\n" + "echo foreign\n"
	require.NoError(t, os.WriteFile(hookPath, []byte(existing), 0o755))

	changed, err := UninstallHook(repoRoot)
	require.NoError(t, err)
	assert.True(t, changed)

	data, err := os.ReadFile(hookPath)
	require.NoError(t, err)
	content := string(data)
	assert.NotContains(t, content, hookMarker)
	assert.NotContains(t, content, hookMarkerEnd)
	assert.NotContains(t, content, "ENVGUARD_BIN=")
	assert.Contains(t, content, "echo foreign")
}

func TestFindRepoRootSupportsDotGitFile(t *testing.T) {
	repoRoot, gitDir := initTestRepoWithGitFile(t)

	nested := filepath.Join(repoRoot, "nested", "deeper")
	require.NoError(t, os.MkdirAll(nested, 0o755))

	foundRoot, err := FindRepoRoot(nested)
	require.NoError(t, err)
	assert.Equal(t, repoRoot, foundRoot)

	foundGitDir, err := GitDir(repoRoot)
	require.NoError(t, err)
	assert.Equal(t, gitDir, foundGitDir)
}

func TestHookPathSupportsDotGitFile(t *testing.T) {
	repoRoot, gitDir := initTestRepoWithGitFile(t)

	hookPath, err := HookPath(repoRoot)
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(gitDir, "hooks", "pre-commit"), hookPath)
}

func initTestRepo(t *testing.T) string {
	t.Helper()
	repoRoot := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(repoRoot, ".git", "hooks"), 0o755))
	return repoRoot
}

func initTestRepoWithGitFile(t *testing.T) (string, string) {
	t.Helper()
	base := t.TempDir()
	repoRoot := filepath.Join(base, "worktree")
	gitDir := filepath.Join(base, "actual-git-dir")
	require.NoError(t, os.MkdirAll(filepath.Join(repoRoot), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(gitDir, "hooks"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(repoRoot, ".git"), []byte("gitdir: ../actual-git-dir\n"), 0o644))
	return repoRoot, gitDir
}
