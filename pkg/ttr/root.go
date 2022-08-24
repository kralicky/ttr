package ttr

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kralicky/ttr/pkg/config"
	"github.com/kralicky/ttr/pkg/game"
	"github.com/kralicky/ttr/pkg/ttr/commands"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
func BuildRootCmd() *cobra.Command {

	rootCmd := &cobra.Command{
		Use:   "ttr",
		Short: "A brief description of your application",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			dataDir, err := game.UpsertDataDir()
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			configFile := filepath.Join(dataDir, "cli-config.yaml")
			return config.Load(configFile)
		},
	}

	accountsCmd := commands.BuildAccountsCmd()
	accountsCmd.AddCommand(commands.BuildAddCmd())
	accountsCmd.AddCommand(commands.BuildListCmd())
	accountsCmd.AddCommand(commands.BuildRmCmd())
	accountsCmd.AddCommand(commands.BuildEditCmd())

	rootCmd.AddCommand(accountsCmd)
	rootCmd.AddCommand(commands.BuildLaunchCmd())
	rootCmd.AddCommand(commands.BuildConfigCmd())
	//+cobra:subcommands

	return rootCmd
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := BuildRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}
