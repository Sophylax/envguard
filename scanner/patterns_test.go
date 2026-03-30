package scanner

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuiltInPatternsAgainstFixtures(t *testing.T) {
	rules, err := BuiltInPatterns()
	require.NoError(t, err)

	fixtures, err := loadPatternFixtures(filepath.Join("..", "testdata", "fake_keys.txt"))
	require.NoError(t, err)

	for _, rule := range rules {
		samples, ok := fixtures[rule.Name]
		require.Truef(t, ok, "missing fixtures for %s", rule.Name)

		var positives, negatives []string
		for _, sample := range samples {
			if sample.kind == "positive" {
				positives = append(positives, sample.value)
			} else {
				negatives = append(negatives, sample.value)
			}
		}

		require.Lenf(t, positives, 2, "need 2 positives for %s", rule.Name)
		require.Lenf(t, negatives, 1, "need 1 negative for %s", rule.Name)

		for _, positive := range positives {
			assert.Truef(t, rule.Regex.MatchString(positive), "expected %s to match %s", positive, rule.Name)
		}
		for _, negative := range negatives {
			assert.Falsef(t, rule.Regex.MatchString(negative), "expected %s not to match %s", negative, rule.Name)
		}
	}
}

type patternFixture struct {
	kind  string
	value string
}

func loadPatternFixtures(path string) (map[string][]patternFixture, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open fixture file %s: %w", path, err)
	}
	defer file.Close()

	out := map[string][]patternFixture{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		var name, kind, value string
		n, err := fmt.Sscanf(line, "%[^|]|%[^|]|%s", &name, &kind, &value)
		if err != nil || n != 3 {
			parts := splitFixtureLine(line)
			if len(parts) != 3 {
				return nil, fmt.Errorf("parse fixture line %q: %w", line, err)
			}
			name, kind, value = parts[0], parts[1], parts[2]
		}
		out[name] = append(out[name], patternFixture{kind: kind, value: value})
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan fixture file %s: %w", path, err)
	}
	return out, nil
}

func splitFixtureLine(line string) []string {
	var parts []string
	current := ""
	for _, r := range line {
		if r == '|' && len(parts) < 2 {
			parts = append(parts, current)
			current = ""
			continue
		}
		current += string(r)
	}
	parts = append(parts, current)
	return parts
}
