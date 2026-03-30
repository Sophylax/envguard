package git

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const hookMarker = "# envguard pre-commit hook"
const hookMarkerEnd = "# end envguard pre-commit hook"

// InstallOptions controls how envguard merges with existing hooks.
type InstallOptions struct {
	BinaryPath  string
	Force       bool
	Interactive bool
}

// FindRepoRoot walks upward from startDir until it finds a .git directory.
func FindRepoRoot(startDir string) (string, error) {
	dir, err := filepath.Abs(startDir)
	if err != nil {
		return "", fmt.Errorf("resolve start dir %s: %w", startDir, err)
	}
	for {
		candidate := filepath.Join(dir, ".git")
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return dir, nil
		} else if err != nil && !errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("stat %s: %w", candidate, err)
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("git repository not found from %s", startDir)
		}
		dir = parent
	}
}

// HookPath returns the pre-commit hook path for the repo.
func HookPath(repoRoot string) string {
	return filepath.Join(repoRoot, ".git", "hooks", "pre-commit")
}

// InstallHook installs or updates the envguard pre-commit hook.
func InstallHook(repoRoot string, in io.Reader, out io.Writer, opts InstallOptions) (string, error) {
	hookPath := HookPath(repoRoot)
	hookScript := buildHookScript(opts.BinaryPath)
	existing, err := os.ReadFile(hookPath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("read hook %s: %w", hookPath, err)
	}

	content := string(existing)
	switch {
	case content == "":
		if err := writeHook(hookPath, hookScript); err != nil {
			return "", fmt.Errorf("write new hook %s: %w", hookPath, err)
		}
	case strings.Contains(content, hookMarker):
		if err := writeHook(hookPath, rewriteWithEnvguard(content, opts.BinaryPath)); err != nil {
			return "", fmt.Errorf("overwrite envguard hook %s: %w", hookPath, err)
		}
	default:
		if opts.Force {
			merged := hookScript + "\n" + content
			if err := writeHook(hookPath, merged); err != nil {
				return "", fmt.Errorf("merge envguard hook into %s: %w", hookPath, err)
			}
			return hookPath, nil
		}
		if !opts.Interactive {
			return "", fmt.Errorf("foreign pre-commit hook exists at %s; rerun with --yes to prepend envguard non-interactively", hookPath)
		}
		fmt.Fprintln(out, "Warning: existing pre-commit hook found.")
		fmt.Fprint(out, "Prepend envguard to existing hook? [y/N]: ")
		ok, err := confirm(in)
		if err != nil {
			return "", fmt.Errorf("read confirmation: %w", err)
		}
		if !ok {
			return "", fmt.Errorf("installation cancelled")
		}
		merged := hookScript + "\n" + content
		if err := writeHook(hookPath, merged); err != nil {
			return "", fmt.Errorf("merge envguard hook into %s: %w", hookPath, err)
		}
	}

	return hookPath, nil
}

// UninstallHook removes envguard from the pre-commit hook without disturbing foreign content.
func UninstallHook(repoRoot string) (bool, error) {
	hookPath := HookPath(repoRoot)
	data, err := os.ReadFile(hookPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, fmt.Errorf("read hook %s: %w", hookPath, err)
	}

	content := strings.TrimSpace(removeEnvguardBlock(string(data)))
	if content == "" {
		if err := os.Remove(hookPath); err != nil {
			return false, fmt.Errorf("remove hook %s: %w", hookPath, err)
		}
		return true, nil
	}

	if err := writeHook(hookPath, content+"\n"); err != nil {
		return false, fmt.Errorf("rewrite hook %s: %w", hookPath, err)
	}
	return true, nil
}

func writeHook(path string, content string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create hook dir for %s: %w", path, err)
	}
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"+strings.TrimPrefix(content, "#!/bin/sh\n")), 0o755); err != nil {
		return fmt.Errorf("write hook file %s: %w", path, err)
	}
	return nil
}

func buildHookScript(binaryPath string) string {
	return fmt.Sprintf(`%s
ENVGUARD_BIN=%s
if command -v envguard >/dev/null 2>&1; then
  envguard check
elif [ -n "$ENVGUARD_BIN" ] && [ -x "$ENVGUARD_BIN" ]; then
  "$ENVGUARD_BIN" check
else
  echo "envguard binary not found. Re-run 'envguard install' to refresh the hook."
  exit 1
fi
%s
`, hookMarker, shellQuote(binaryPath), hookMarkerEnd)
}

func shellQuote(value string) string {
	if value == "" {
		return `""`
	}
	return "'" + strings.ReplaceAll(value, "'", `'"'"'`) + "'"
}

func confirm(in io.Reader) (bool, error) {
	reader := bufio.NewReader(in)
	line, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return false, fmt.Errorf("read confirmation input: %w", err)
	}
	answer := strings.ToLower(strings.TrimSpace(line))
	return answer == "y" || answer == "yes", nil
}

func rewriteWithEnvguard(content string, binaryPath string) string {
	if binaryPath == "" {
		binaryPath = extractEnvguardBinaryPath(content)
	}
	hookScript := buildHookScript(binaryPath)
	remaining := strings.TrimSpace(removeEnvguardBlock(content))
	if remaining == "" {
		return hookScript
	}
	return hookScript + "\n" + remaining + "\n"
}

func extractEnvguardBinaryPath(content string) string {
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "ENVGUARD_BIN=") {
			continue
		}
		value := strings.TrimPrefix(trimmed, "ENVGUARD_BIN=")
		if unquoted, ok := unquoteShellString(value); ok {
			return unquoted
		}
	}
	return ""
}

func unquoteShellString(value string) (string, bool) {
	if len(value) >= 2 && value[0] == '\'' && value[len(value)-1] == '\'' {
		unquoted := strings.TrimSuffix(strings.TrimPrefix(value, "'"), "'")
		return strings.ReplaceAll(unquoted, `'"'"'`, `'`), true
	}
	if value == `""` {
		return "", true
	}
	return "", false
}

func removeEnvguardBlock(content string) string {
	lines := strings.Split(content, "\n")
	var out []string
	skip := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == hookMarker {
			skip = true
			continue
		}
		if skip {
			if trimmed == hookMarkerEnd {
				skip = false
				continue
			}
			if trimmed == "" {
				skip = false
				continue
			}
			if strings.HasPrefix(trimmed, "ENVGUARD_BIN=") ||
				trimmed == "if command -v envguard >/dev/null 2>&1; then" ||
				trimmed == "envguard check" ||
				trimmed == `elif [ -n "$ENVGUARD_BIN" ] && [ -x "$ENVGUARD_BIN" ]; then` ||
				trimmed == `"$ENVGUARD_BIN" check` ||
				trimmed == "else" ||
				trimmed == `echo "envguard binary not found. Re-run 'envguard install' to refresh the hook."` ||
				trimmed == `echo "envguard binary not found in PATH"` ||
				trimmed == "exit 1" ||
				trimmed == "fi" {
				continue
			}
			skip = false
		}
		if trimmed == "#!/bin/sh" {
			continue
		}
		out = append(out, line)
	}
	return strings.TrimSpace(strings.Join(out, "\n"))
}
