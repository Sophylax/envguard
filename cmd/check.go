package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sophylax/envguard/allowlist"
	"github.com/sophylax/envguard/config"
	envgit "github.com/sophylax/envguard/git"
	"github.com/sophylax/envguard/reporter"
	"github.com/sophylax/envguard/scanner"
	"github.com/spf13/cobra"
)

var envgitStagedFiles = envgit.StagedFiles

func newCheckCommand() *cobra.Command {
	var scanAll bool
	var jsonOutput bool
	var severityFilter string

	cmd := &cobra.Command{
		Use:   "check [path]",
		Short: "Scan staged files or a provided path for secrets",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("get working directory: %w", err)
			}
			cfg, _, err := config.Load(cwd)
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			repoRoot, err := envgit.FindRepoRoot(cwd)
			if err != nil {
				repoRoot = cwd
			}
			allow, _, err := allowlist.Load(repoRoot)
			if err != nil {
				return fmt.Errorf("load allowlist: %w", err)
			}

			paths, scope, pathWarnings, err := resolveScanPaths(args, scanAll)
			if err != nil {
				return fmt.Errorf("resolve scan paths: %w", err)
			}

			engine, err := scanner.NewEngine(cfg, allow)
			if err != nil {
				return fmt.Errorf("create scanner: %w", err)
			}
			findings, err := engine.ScanPaths(paths)
			if err != nil {
				return fmt.Errorf("scan paths: %w", err)
			}
			findings = filterBySeverity(findings, severityFilter)

			if jsonOutput {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				if err := enc.Encode(findings); err != nil {
					return fmt.Errorf("encode findings json: %w", err)
				}
			} else {
				if err := reporter.PrintScanHeader(cmd.OutOrStdout(), len(paths), scope); err != nil {
					return fmt.Errorf("print header: %w", err)
				}
				warnings := append(pathWarnings, engine.Warnings()...)
				if err := reporter.PrintWarnings(cmd.OutOrStdout(), warnings); err != nil {
					return fmt.Errorf("print warnings: %w", err)
				}
				if len(findings) == 0 {
					if err := reporter.PrintSummary(cmd.OutOrStdout(), findings); err != nil {
						return fmt.Errorf("print summary: %w", err)
					}
				} else {
					if err := reporter.PrintFindings(cmd.OutOrStdout(), findings); err != nil {
						return fmt.Errorf("print findings: %w", err)
					}
					if err := reporter.PrintSummary(cmd.OutOrStdout(), findings); err != nil {
						return fmt.Errorf("print summary: %w", err)
					}
				}
			}

			if len(findings) > 0 {
				return ErrFindings
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&scanAll, "all", false, "scan the entire working tree")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "output findings as JSON")
	cmd.Flags().StringVar(&severityFilter, "severity", "", "filter output by severity: HIGH|MEDIUM|LOW")
	return cmd
}

func resolveScanPaths(args []string, scanAll bool) ([]string, string, []string, error) {
	if len(args) == 1 {
		return []string{args[0]}, "path entries", nil, nil
	}
	if scanAll {
		return []string{"."}, "working tree files", nil, nil
	}
	staged, err := envgitStagedFiles()
	if err != nil {
		return nil, "", nil, fmt.Errorf("list staged files: %w", err)
	}
	if len(staged) == 0 {
		return []string{}, "staged files", nil, nil
	}
	paths := make([]string, 0, len(staged))
	warnings := make([]string, 0)
	for _, path := range staged {
		cleanPath := filepath.Clean(path)
		if _, err := os.Stat(cleanPath); err != nil {
			if os.IsNotExist(err) {
				warnings = append(warnings, fmt.Sprintf("skipping %s: path no longer exists", cleanPath))
				continue
			}
			return nil, "", nil, fmt.Errorf("stat staged path %s: %w", cleanPath, err)
		}
		paths = append(paths, cleanPath)
	}
	return paths, "staged files", warnings, nil
}

func filterBySeverity(findings []scanner.Finding, severity string) []scanner.Finding {
	if severity == "" {
		return findings
	}
	severity = strings.ToUpper(severity)
	filtered := make([]scanner.Finding, 0, len(findings))
	for _, finding := range findings {
		if strings.ToUpper(finding.Severity) == severity {
			filtered = append(filtered, finding)
		}
	}
	return filtered
}
