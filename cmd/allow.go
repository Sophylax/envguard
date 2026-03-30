package cmd

import (
	"fmt"
	"os"

	"github.com/sophylax/envguard/allowlist"
	envgit "github.com/sophylax/envguard/git"
	"github.com/spf13/cobra"
)

func newAllowCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "allow <fingerprint>",
		Short: "Add a finding fingerprint to the repo allowlist",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("get working directory: %w", err)
			}
			repoRoot, err := envgit.FindRepoRoot(cwd)
			if err != nil {
				return fmt.Errorf("find git repository: %w", err)
			}
			if _, err := allowlist.Add(repoRoot, args[0]); err != nil {
				return fmt.Errorf("add fingerprint to allowlist: %w", err)
			}
			if _, err := fmt.Fprintf(cmd.OutOrStdout(), "Fingerprint %s added to .envguard-ignore\n", args[0]); err != nil {
				return fmt.Errorf("write allow confirmation: %w", err)
			}
			return nil
		},
	}
}
