package commands

import (
	"fmt"

	"github.com/kralicky/ttr/pkg/auth"
	"github.com/kralicky/ttr/pkg/config"
	"github.com/spf13/cobra"
)

// RmCmd represents the rm command
func BuildRmCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "rm <username>",
		Aliases: []string{"remove", "delete"},
		Short:   "Remove a stored account",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !config.AccountExists(args[0]) {
				return fmt.Errorf("account %s does not exist", args[0])
			}
			if err := auth.DeleteAccountPassword(args[0]); err != nil {
				return fmt.Errorf("failed to delete credentials: %w", err)
			}
			config.DeleteAccount(args[0])
			if err := config.Save(); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}
			return nil
		},
	}
	return cmd
}
