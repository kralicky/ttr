package commands

import (
	"github.com/spf13/cobra"
)

// ConfigCmd represents the config command
func BuildConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "A brief description of your command",
	}
	return cmd
}
