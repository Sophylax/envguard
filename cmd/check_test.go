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

func chdirForTest(t *testing.T, dir string) {
	t.Helper()
	wd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() {
		require.NoError(t, os.Chdir(wd))
	})
}
