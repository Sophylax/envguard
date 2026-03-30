package scanner

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/sophylax/envguard/config"
)

// PatternRule defines a regex-based detection rule.
type PatternRule struct {
	Name        string
	Regex       *regexp.Regexp
	Severity    string
	Description string
}

var builtInPatternSpecs = []struct {
	name        string
	pattern     string
	severity    string
	description string
}{
	{"AWS Access Key", `AKIA[0-9A-Z]{16}`, "HIGH", "Detects AWS access key identifiers."},
	{"AWS Secret Key", `(?i)aws_secret[^A-Za-z0-9]{0,20}[A-Za-z0-9/+=]{40}`, "HIGH", "Detects 40-character AWS secret values near aws_secret markers."},
	{"Generic API Key", `(?i)(api_key|apikey|api-key)\s*[:=]\s*["']?[A-Za-z0-9_\-]{20,}`, "HIGH", "Detects generic API key assignments."},
	{"Generic Secret", `(?i)(secret|password|passwd|pwd)\s*[:=]\s*["']?\S{8,}`, "MEDIUM", "Detects generic secret or password assignments."},
	{"Private RSA Key", `-----BEGIN (RSA|EC|OPENSSH) PRIVATE KEY-----`, "HIGH", "Detects private key blocks."},
	{"JWT", `eyJ[A-Za-z0-9_-]{10,}\.[A-Za-z0-9_-]{10,}\.`, "MEDIUM", "Detects JWT-like bearer tokens."},
	{"Database URL", `(postgres|mysql|mongodb)://[^:\s]+:[^@\s]+@`, "HIGH", "Detects database connection URLs with inline credentials."},
	{"Slack Token", `xox[baprs]-[A-Za-z0-9-]{10,}`, "HIGH", "Detects Slack tokens."},
	{"GitHub Token", `gh[pousr]_[A-Za-z0-9_]{36,}`, "HIGH", "Detects GitHub personal and app tokens."},
	{"Generic Bearer Token", `[Bb]earer\s+[A-Za-z0-9\-_.]{20,}`, "MEDIUM", "Detects bearer authorization headers."},
	{"Google API Key", `AIza[0-9A-Za-z\-_]{35}`, "HIGH", "Detects Google API keys."},
	{"Stripe Key", `sk_(live|test)_[A-Za-z0-9]{24,}`, "HIGH", "Detects Stripe secret keys."},
	{"Twilio Key", `SK[a-fA-F0-9]{32}`, "HIGH", "Detects Twilio API keys."},
}

// BuiltInPatterns returns the compiled built-in rules.
func BuiltInPatterns() ([]PatternRule, error) {
	rules := make([]PatternRule, 0, len(builtInPatternSpecs))
	for _, spec := range builtInPatternSpecs {
		re, err := regexp.Compile(spec.pattern)
		if err != nil {
			return nil, fmt.Errorf("compile pattern %q: %w", spec.name, err)
		}
		rules = append(rules, PatternRule{
			Name:        spec.name,
			Regex:       re,
			Severity:    spec.severity,
			Description: spec.description,
		})
	}
	return rules, nil
}

// AllPatterns returns built-in rules plus custom config rules.
func AllPatterns(cfg config.Config) ([]PatternRule, error) {
	rules, err := BuiltInPatterns()
	if err != nil {
		return nil, fmt.Errorf("load built-in patterns: %w", err)
	}
	for _, custom := range cfg.CustomPatterns {
		re, err := regexp.Compile(custom.Pattern)
		if err != nil {
			return nil, fmt.Errorf("compile custom pattern %q: %w", custom.Name, err)
		}
		rules = append(rules, PatternRule{
			Name:        custom.Name,
			Regex:       re,
			Severity:    strings.ToUpper(custom.Severity),
			Description: "Custom user-defined pattern.",
		})
	}
	return rules, nil
}
