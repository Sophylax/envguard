package cmd

import (
	"fmt"
	"os"

	envgit "github.com/sophylax/envguard/git"
	"github.com/spf13/cobra"
)

func newUninstallCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "uninstall",
		Short: "Remove the envguard git pre-commit hook",
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("get working directory: %w", err)
			}
			repoRoot, err := envgit.FindRepoRoot(cwd)
			if err != nil {
				return fmt.Errorf("find git repository: %w", err)
			}
			changed, err := envgit.UninstallHook(repoRoot)
			if err != nil {
				return fmt.Errorf("uninstall hook: %w", err)
			}
			message := "envguard hook was not installed\n"
			if changed {
				message = "envguard hook removed\n"
			}
			if _, err := cmd.OutOrStdout().Write([]byte(message)); err != nil {
				return fmt.Errorf("write uninstall confirmation: %w", err)
			}
			return nil
		},
	}
}
