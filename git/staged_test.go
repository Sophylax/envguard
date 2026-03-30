package git

import (
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStagedFiles(t *testing.T) {
	original := execCommand
	execCommand = fakeExecCommand
	defer func() {
		execCommand = original
	}()

	files, err := StagedFiles()
	require.NoError(t, err)
	assert.Equal(t, []string{"cmd/check.go", "scanner/scanner.go"}, files)
}

func fakeExecCommand(command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestHelperProcess", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = append(os.Environ(), "GO_WANT_HELPER_PROCESS=1")
	return cmd
}

func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	_, _ = os.Stdout.WriteString("cmd/check.go\nscanner/scanner.go\n")
	os.Exit(0)
}
