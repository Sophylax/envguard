package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

// ErrFindings signals that a scan found reportable secrets and should exit non-zero without extra stderr noise.
var ErrFindings = errors.New("findings detected")

// Execute runs the root command tree.
func Execute(version string) error {
	root := newRootCommand(version)
	if err := root.Execute(); err != nil {
		return fmt.Errorf("execute command: %w", err)
	}
	return nil
}

func newRootCommand(version string) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:           "envguard",
		Short:         "Zero-config pre-commit secret scanner",
		SilenceErrors: true,
		SilenceUsage:  true,
	}
	rootCmd.Version = version
	rootCmd.SetVersionTemplate("{{.Version}}\n")

	rootCmd.AddCommand(newCheckCommand())
	rootCmd.AddCommand(newInstallCommand())
	rootCmd.AddCommand(newUninstallCommand())
	rootCmd.AddCommand(newAllowCommand())
	rootCmd.AddCommand(newVersionCommand(version))
	return rootCmd
}

func newVersionCommand(version string) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print envguard version",
		RunE: func(cmd *cobra.Command, args []string) error {
			if _, err := cmd.OutOrStdout().Write([]byte(version + "\n")); err != nil {
				return fmt.Errorf("write version: %w", err)
			}
			return nil
		},
	}
}
