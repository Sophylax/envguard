package cmd

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/sophylax/envguard/scanner"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckCommandJSONOutputIsParseable(t *testing.T) {
	tempDir := t.TempDir()
	chdirForTest(t, tempDir)

	target := filepath.Join(tempDir, "secret.js")
	require.NoError(t, os.WriteFile(target, []byte("const key = \"AKIA1234567890ABCDEF\";\n"), 0o644))

	cmd := newCheckCommand()
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	// Keep Cobra from mixing usage/error text into stdout so --json output stays parseable in tests.
	output := &bytes.Buffer{}
	cmd.SetOut(output)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{target, "--json"})

	err := cmd.Execute()
	require.ErrorIs(t, err, ErrFindings)

	var findings []scanner.Finding
	require.NoError(t, json.Unmarshal(output.Bytes(), &findings))
	require.Len(t, findings, 1)
	assert.Equal(t, "HIGH", findings[0].Severity)
	assert.Equal(t, "AWS Access Key", findings[0].RuleName)
}

func TestCheckCommandSeverityFilterAppliesToJSONOutput(t *testing.T) {
	tempDir := t.TempDir()
	chdirForTest(t, tempDir)

	target := filepath.Join(tempDir, "mixed.txt")
	require.NoError(t, os.WriteFile(target, []byte("const key = \"AKIA1234567890ABCDEF\";\nconst token = \"kR9mXp2Qa7Vz1Ld8Hs4Ty6Nu0We3\";\n"), 0o644))

	cmd := newCheckCommand()
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	// Keep Cobra from mixing usage/error text into stdout so --json output stays parseable in tests.
	output := &bytes.Buffer{}
	cmd.SetOut(output)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{target, "--json", "--severity", "HIGH"})

	err := cmd.Execute()
	require.ErrorIs(t, err, ErrFindings)

	var findings []scanner.Finding
	require.NoError(t, json.Unmarshal(output.Bytes(), &findings))
	require.Len(t, findings, 1)
	assert.Equal(t, "HIGH", findings[0].Severity)
	assert.Equal(t, "AWS Access Key", findings[0].RuleName)
}

func TestCheckCommandSkipsMissingStagedFiles(t *testing.T) {
	tempDir := t.TempDir()
	chdirForTest(t, tempDir)

	target := filepath.Join(tempDir, "secret.js")
	require.NoError(t, os.WriteFile(target, []byte("const key = \"AKIA1234567890ABCDEF\";\n"), 0o644))

	original := envgitStagedFiles
	envgitStagedFiles = func() ([]string, error) {
		return []string{"deleted.txt", "secret.js"}, nil
	}
	t.Cleanup(func() {
		envgitStagedFiles = original
	})

	cmd := newCheckCommand()
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	output := &bytes.Buffer{}
	cmd.SetOut(output)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	require.ErrorIs(t, err, ErrFindings)
	assert.Contains(t, output.String(), "[envguard] scanning 1 staged files...")
	assert.Contains(t, output.String(), "[envguard] warning: skipping deleted.txt: path no longer exists")
	assert.Contains(t, output.String(), "AWS Access Key")
}

func TestCheckCommandExplicitMissingPathStillErrors(t *testing.T) {
	tempDir := t.TempDir()
	chdirForTest(t, tempDir)

	missing := filepath.Join(tempDir, "missing.txt")

	cmd := newCheckCommand()
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{missing})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "scan paths")
	assert.Contains(t, err.Error(), "stat path")
}

func TestCheckCommandFingerprintStableAcrossWorkingDirectories(t *testing.T) {
	repoRoot := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(repoRoot, ".git", "hooks"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(repoRoot, "nested"), 0o755))

	target := filepath.Join(repoRoot, "secret.js")
	require.NoError(t, os.WriteFile(target, []byte("const key = \"AKIA1234567890ABCDEF\";\n"), 0o644))

	runCheck := func(t *testing.T, workdir string, arg string) scanner.Finding {
		t.Helper()
		chdirForTest(t, workdir)

		cmd := newCheckCommand()
		cmd.SilenceErrors = true
		cmd.SilenceUsage = true
		output := &bytes.Buffer{}
		cmd.SetOut(output)
		cmd.SetErr(io.Discard)
		cmd.SetArgs([]string{arg, "--json"})

		err := cmd.Execute()
		require.ErrorIs(t, err, ErrFindings)

		var findings []scanner.Finding
		require.NoError(t, json.Unmarshal(output.Bytes(), &findings))
		require.Len(t, findings, 1)
		return findings[0]
	}

	rootFinding := runCheck(t, repoRoot, "secret.js")
	nestedFinding := runCheck(t, filepath.Join(repoRoot, "nested"), "../secret.js")

	assert.Equal(t, "secret.js", rootFinding.File)
	assert.Equal(t, "secret.js", nestedFinding.File)
	assert.Equal(t, rootFinding.Fingerprint, nestedFinding.Fingerprint)
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
