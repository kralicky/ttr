package ttr

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kralicky/ttr/pkg/config"
	"github.com/kralicky/ttr/pkg/game"
	"github.com/kralicky/ttr/pkg/ttr/commands"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
func BuildRootCmd() *cobra.Command {
	var logLevel string
	rootCmd := &cobra.Command{
		Use:          "ttr",
		Short:        "TTR CLI Launcher",
		SilenceUsage: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			level, err := logrus.ParseLevel(logLevel)
			if err != nil {
				return err
			}
			logrus.SetLevel(level)

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
	accountsCmd.AddCommand(commands.BuildTwoFactorAuthCmd())

	rootCmd.AddCommand(accountsCmd)
	rootCmd.AddCommand(commands.BuildLaunchCmd())
	rootCmd.AddCommand(commands.BuildConfigCmd())
	rootCmd.AddCommand(commands.BuildDirCmd())
	rootCmd.AddCommand(commands.BuildMultitoonCmd())
	rootCmd.AddCommand(commands.BuildStatusCmd())
	//+cobra:subcommands

	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "Log level (debug, info, warn, error)")
	return rootCmd
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := BuildRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}
