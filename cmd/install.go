package cmd

import (
	"fmt"
	"os"

	envgit "github.com/sophylax/envguard/git"
	"github.com/spf13/cobra"
)

func newInstallCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "install",
		Short: "Install the git pre-commit hook",
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("get working directory: %w", err)
			}
			repoRoot, err := envgit.FindRepoRoot(cwd)
			if err != nil {
				return fmt.Errorf("find git repository: %w", err)
			}
			hookPath, err := envgit.InstallHook(repoRoot, cmd.InOrStdin(), cmd.OutOrStdout())
			if err != nil {
				return fmt.Errorf("install hook: %w", err)
			}
			if _, err := fmt.Fprintf(cmd.OutOrStdout(), "envguard hook installed at %s\n", hookPath); err != nil {
				return fmt.Errorf("write install confirmation: %w", err)
			}
			return nil
		},
	}
}
