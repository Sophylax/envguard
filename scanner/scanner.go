package scanner

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"unicode"

	"github.com/sophylax/envguard/allowlist"
	"github.com/sophylax/envguard/config"
)

var tokenSplitRE = regexp.MustCompile("[\\s=:\"'`,;()]+")

// Finding describes a detected secret candidate.
type Finding struct {
	File        string `json:"file"`
	Line        int    `json:"line"`
	Column      int    `json:"column"`
	Value       string `json:"value"`
	RuleName    string `json:"rule_name"`
	Severity    string `json:"severity"`
	Fingerprint string `json:"fingerprint"`
	Source      string `json:"source"`
}

// Engine scans files using built-in pattern and entropy detectors.
type Engine struct {
	cfg      config.Config
	allow    allowlist.Set
	patterns []PatternRule
	warnings []string
}

// NewEngine constructs a scanner engine.
func NewEngine(cfg config.Config, allow allowlist.Set) (*Engine, error) {
	patterns, err := AllPatterns(cfg)
	if err != nil {
		return nil, fmt.Errorf("build pattern set: %w", err)
	}
	if allow == nil {
		allow = allowlist.Set{}
	}
	return &Engine{cfg: cfg, allow: allow, patterns: patterns}, nil
}

// Warnings returns non-fatal scan warnings, such as skipped oversized files.
func (e *Engine) Warnings() []string {
	return append([]string(nil), e.warnings...)
}

// ScanPaths scans one or more files or directories recursively.
func (e *Engine) ScanPaths(paths []string) ([]Finding, error) {
	expanded, err := e.expandPaths(paths)
	if err != nil {
		return nil, fmt.Errorf("expand scan paths: %w", err)
	}

	var findings []Finding
	for _, path := range expanded {
		fileFindings, err := e.scanFile(path)
		if err != nil {
			return nil, fmt.Errorf("scan file %s: %w", path, err)
		}
		findings = append(findings, fileFindings...)
	}

	sort.Slice(findings, func(i, j int) bool {
		if findings[i].Severity != findings[j].Severity {
			return severityRank(findings[i].Severity) < severityRank(findings[j].Severity)
		}
		if findings[i].File != findings[j].File {
			return findings[i].File < findings[j].File
		}
		if findings[i].Line != findings[j].Line {
			return findings[i].Line < findings[j].Line
		}
		return findings[i].Column < findings[j].Column
	})

	return dedupeFindings(findings), nil
}

func (e *Engine) expandPaths(paths []string) ([]string, error) {
	seen := map[string]struct{}{}
	var files []string
	for _, input := range paths {
		abs, err := filepath.Abs(input)
		if err != nil {
			return nil, fmt.Errorf("resolve path %s: %w", input, err)
		}
		info, err := os.Stat(abs)
		if err != nil {
			return nil, fmt.Errorf("stat path %s: %w", abs, err)
		}
		if info.IsDir() {
			if err := filepath.WalkDir(abs, func(path string, d fs.DirEntry, walkErr error) error {
				if walkErr != nil {
					return fmt.Errorf("walk %s: %w", path, walkErr)
				}
				if d.IsDir() {
					return nil
				}
				if _, ok := seen[path]; ok {
					return nil
				}
				if e.shouldExclude(path) {
					return nil
				}
				seen[path] = struct{}{}
				files = append(files, path)
				return nil
			}); err != nil {
				return nil, fmt.Errorf("walk directory %s: %w", abs, err)
			}
			continue
		}
		if e.shouldExclude(abs) {
			continue
		}
		if _, ok := seen[abs]; ok {
			continue
		}
		seen[abs] = struct{}{}
		files = append(files, abs)
	}
	sort.Strings(files)
	return files, nil
}

func (e *Engine) scanFile(path string) ([]Finding, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("stat file %s: %w", path, err)
	}
	if limit := int64(e.cfg.MaxFileSizeKB) * 1024; limit > 0 && info.Size() > limit {
		e.warnings = append(e.warnings, fmt.Sprintf("skipping %s: file size %d bytes exceeds limit %d KB", path, info.Size(), e.cfg.MaxFileSizeKB))
		return nil, nil
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open file %s: %w", path, err)
	}
	defer file.Close()

	relative, err := filepath.Rel(mustGetwd(), path)
	if err != nil {
		relative = path
	}
	relative = filepath.ToSlash(relative)

	var findings []Finding
	if envFilePattern.MatchString(filepath.Base(path)) {
		finding := newFinding(relative, 1, 1, filepath.Base(path), ".env file staged", "HIGH", "pattern")
		if !e.allow.Contains(finding.Fingerprint) {
			findings = append(findings, finding)
		}
	}

	scanner := bufio.NewScanner(file)
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()
		patternFindings, spans := e.scanPatterns(relative, lineNumber, line)
		findings = append(findings, patternFindings...)
		if e.cfg.AllowTestFixtures && isTestdataPath(relative) {
			continue
		}
		findings = append(findings, e.scanEntropy(relative, lineNumber, line, spans)...)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan lines in %s: %w", path, err)
	}
	return findings, nil
}

func (e *Engine) scanPatterns(file string, lineNumber int, line string) ([]Finding, [][2]int) {
	var findings []Finding
	var spans [][2]int
	for _, rule := range e.patterns {
		locs := rule.Regex.FindAllStringIndex(line, -1)
		for _, loc := range locs {
			raw := line[loc[0]:loc[1]]
			finding := newFinding(file, lineNumber, loc[0]+1, raw, rule.Name, rule.Severity, "pattern")
			if e.allow.Contains(finding.Fingerprint) {
				continue
			}
			findings = append(findings, finding)
			spans = append(spans, [2]int{loc[0], loc[1]})
		}
	}
	return findings, spans
}

func (e *Engine) scanEntropy(file string, lineNumber int, line string, patternSpans [][2]int) []Finding {
	tokens := tokenSplitRE.Split(line, -1)
	var findings []Finding
	offset := 0
	for _, token := range tokens {
		if token == "" {
			continue
		}
		idx := strings.Index(line[offset:], token)
		column := offset + idx + 1
		offset = column - 1 + len(token)
		start := column - 1
		end := start + len(token)

		if len(token) < e.cfg.MinLength {
			continue
		}
		if overlapsPattern(start, end, patternSpans) {
			continue
		}
		entropy := ShannonEntropy(token)
		if isDictionaryLooking(token) && entropy < 3.2 {
			continue
		}
		if entropy < e.cfg.EntropyThreshold {
			continue
		}

		finding := newFinding(file, lineNumber, column, token, fmt.Sprintf("High Entropy String|%.2f", entropy), "MEDIUM", "entropy")
		if e.allow.Contains(finding.Fingerprint) {
			continue
		}
		findings = append(findings, finding)
	}
	return findings
}

func overlapsPattern(start int, end int, spans [][2]int) bool {
	for _, span := range spans {
		if start < span[1] && end > span[0] {
			return true
		}
	}
	return false
}

func newFinding(file string, line int, column int, rawValue string, ruleName string, severity string, source string) Finding {
	return Finding{
		File:        file,
		Line:        line,
		Column:      column,
		Value:       truncateValue(rawValue),
		RuleName:    ruleName,
		Severity:    strings.ToUpper(severity),
		Fingerprint: Fingerprint(file, line, rawValue),
		Source:      source,
	}
}

func truncateValue(raw string) string {
	if len(raw) <= 6 {
		return raw
	}
	return raw[:6] + "..."
}

func isDictionaryLooking(token string) bool {
	hasLetter := false
	for _, r := range token {
		if unicode.IsLetter(r) {
			hasLetter = true
		}
		if !(unicode.IsLower(r) || r == '-' || r == '_') {
			return false
		}
	}
	return hasLetter
}

func severityRank(severity string) int {
	switch strings.ToUpper(severity) {
	case "HIGH":
		return 0
	case "MEDIUM":
		return 1
	default:
		return 2
	}
}

func dedupeFindings(findings []Finding) []Finding {
	seen := map[string]struct{}{}
	out := make([]Finding, 0, len(findings))
	for _, finding := range findings {
		key := fmt.Sprintf("%s:%d:%d:%s:%s", finding.File, finding.Line, finding.Column, finding.RuleName, finding.Source)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, finding)
	}
	return out
}

func (e *Engine) shouldExclude(path string) bool {
	ext := filepath.Ext(path)
	for _, excluded := range e.cfg.ExcludeExtensions {
		if ext == excluded {
			return true
		}
	}

	normalized := filepath.ToSlash(path)
	for _, pattern := range e.cfg.ExcludePaths {
		if matchGlob(pattern, normalized) {
			return true
		}
	}
	return false
}

var envFilePattern = regexp.MustCompile(`^\.env(\.\w+)?$`)

func isTestdataPath(path string) bool {
	return strings.Contains(filepath.ToSlash(path), "/testdata/") || strings.HasPrefix(filepath.ToSlash(path), "testdata/")
}

func matchGlob(pattern string, target string) bool {
	pattern = filepath.ToSlash(pattern)
	target = filepath.ToSlash(target)
	if strings.HasPrefix(pattern, "**/") {
		if ok, _ := filepath.Match(strings.TrimPrefix(pattern, "**/"), filepath.Base(target)); ok {
			return true
		}
	}
	if strings.HasSuffix(pattern, "/**") {
		prefix := strings.TrimSuffix(pattern, "/**")
		return strings.Contains(target, prefix+"/") || strings.HasSuffix(target, prefix)
	}
	ok, _ := filepath.Match(pattern, target)
	if ok {
		return true
	}
	ok, _ = filepath.Match(pattern, filepath.Base(target))
	return ok
}

func mustGetwd() string {
	wd, err := os.Getwd()
	if err != nil {
		return "."
	}
	return wd
}
