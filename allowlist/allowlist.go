package allowlist

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const fileName = ".envguard-ignore"

// Set stores allowlisted fingerprints for quick membership checks.
type Set map[string]struct{}

// Load reads the repo allowlist file. Missing files return an empty set.
func Load(repoRoot string) (Set, string, error) {
	path := filepath.Join(repoRoot, fileName)
	file, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Set{}, path, nil
		}
		return nil, "", fmt.Errorf("open allowlist %s: %w", path, err)
	}
	defer file.Close()

	set := Set{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		set[line] = struct{}{}
	}
	if err := scanner.Err(); err != nil {
		return nil, "", fmt.Errorf("scan allowlist %s: %w", path, err)
	}
	return set, path, nil
}

// Contains reports whether the fingerprint is allowlisted.
func (s Set) Contains(fingerprint string) bool {
	_, ok := s[fingerprint]
	return ok
}

// Add appends a fingerprint to the repo allowlist if it is not already present.
func Add(repoRoot, fingerprint string) (string, error) {
	set, path, err := Load(repoRoot)
	if err != nil {
		return "", fmt.Errorf("load allowlist: %w", err)
	}
	if set.Contains(fingerprint) {
		return path, nil
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return "", fmt.Errorf("open allowlist for append %s: %w", path, err)
	}
	defer file.Close()

	if _, err := fmt.Fprintln(file, fingerprint); err != nil {
		return "", fmt.Errorf("append fingerprint to %s: %w", path, err)
	}
	return path, nil
}
