package reporter

import (
	"fmt"
	"io"
	"strings"

	"github.com/fatih/color"
	"github.com/sophylax/envguard/scanner"
)

var (
	highColor   = color.New(color.FgRed, color.Bold)
	mediumColor = color.New(color.FgYellow)
	lowColor    = color.New(color.FgCyan)
	cleanColor  = color.New(color.FgGreen)
)

// PrintScanHeader prints the scan start line.
func PrintScanHeader(w io.Writer, count int, scope string) error {
	if _, err := fmt.Fprintf(w, "[envguard] scanning %d %s...", count, scope); err != nil {
		return fmt.Errorf("write scan header: %w", err)
	}
	return nil
}

// PrintWarnings prints non-fatal warnings.
func PrintWarnings(w io.Writer, warnings []string) error {
	for _, warning := range warnings {
		if _, err := fmt.Fprintf(w, "\n[envguard] warning: %s", warning); err != nil {
			return fmt.Errorf("write warning: %w", err)
		}
	}
	return nil
}

// PrintFindings renders findings to the terminal.
func PrintFindings(w io.Writer, findings []scanner.Finding) error {
	for _, finding := range findings {
		severity := finding.Severity
		switch strings.ToUpper(severity) {
		case "HIGH":
			severity = highColor.Sprintf("%-6s", severity)
		case "MEDIUM":
			severity = mediumColor.Sprintf("%-6s", severity)
		default:
			severity = lowColor.Sprintf("%-6s", severity)
		}

		ruleName := finding.RuleName
		source := finding.Source
		if finding.Source == "entropy" && strings.Contains(finding.RuleName, "|") {
			parts := strings.SplitN(finding.RuleName, "|", 2)
			ruleName = parts[0]
			source = "entropy: " + parts[1]
		}

		if _, err := fmt.Fprintf(w, "\n\n  %s %s:%d:%d    %s  (%s)\n", severity, finding.File, finding.Line, finding.Column, ruleName, source); err != nil {
			return fmt.Errorf("write finding header: %w", err)
		}
		if _, err := fmt.Fprintf(w, "         fingerprint: %s\n", finding.Fingerprint); err != nil {
			return fmt.Errorf("write finding fingerprint: %w", err)
		}
		if _, err := fmt.Fprintf(w, "         value: %s\n", finding.Value); err != nil {
			return fmt.Errorf("write finding value: %w", err)
		}
	}
	return nil
}

// PrintSummary renders the final result line.
func PrintSummary(w io.Writer, findings []scanner.Finding) error {
	if len(findings) == 0 {
		if _, err := fmt.Fprintf(w, " %s clean\n", cleanColor.Sprint("✓")); err != nil {
			return fmt.Errorf("write clean summary: %w", err)
		}
		return nil
	}
	if _, err := fmt.Fprintf(w, "\n%d findings. Commit blocked.\nRun: envguard allow <fingerprint>  to whitelist false positives.\n", len(findings)); err != nil {
		return fmt.Errorf("write blocked summary: %w", err)
	}
	return nil
}
