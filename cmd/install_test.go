package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	envgit "github.com/sophylax/envguard/git"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsInteractiveInputWithDevNull(t *testing.T) {
	file, err := os.Open(os.DevNull)
	if err != nil {
		t.Fatalf("open %s: %v", os.DevNull, err)
	}
	defer file.Close()

	if isInteractiveInput(file) {
		t.Fatalf("%s should not be treated as interactive input", os.DevNull)
	}
}

func TestInstallCommandYesMergesForeignHook(t *testing.T) {
	repoRoot := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(repoRoot, ".git", "hooks"), 0o755))

	hookPath := envgit.HookPath(repoRoot)
	require.NoError(t, os.WriteFile(hookPath, []byte("#!/bin/sh\necho foreign\n"), 0o755))

	wd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(repoRoot))
	t.Cleanup(func() {
		require.NoError(t, os.Chdir(wd))
	})

	in, err := os.Open(os.DevNull)
	require.NoError(t, err)
	defer in.Close()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd := newRootCommand("test")
	cmd.SetArgs([]string{"install", "--yes"})
	cmd.SetIn(in)
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	require.NoError(t, cmd.Execute())

	executablePath, err := os.Executable()
	require.NoError(t, err)

	data, err := os.ReadFile(hookPath)
	require.NoError(t, err)
	content := string(data)
	assert.Contains(t, content, "ENVGUARD_BIN='"+executablePath+"'")
	assert.Contains(t, content, "if command -v envguard >/dev/null 2>&1; then")
	assert.Contains(t, content, "elif [ -n \"$ENVGUARD_BIN\" ] && [ -x \"$ENVGUARD_BIN\" ]; then")
	assert.Less(t, strings.Index(content, "envguard check"), strings.Index(content, `"$ENVGUARD_BIN" check`))
	assert.Contains(t, content, "echo foreign")
	assert.Contains(t, stdout.String(), "envguard hook installed at "+hookPath)
	assert.Empty(t, strings.TrimSpace(stderr.String()))
}
