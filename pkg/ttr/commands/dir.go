package commands

import (
	"fmt"
	"os"

	"github.com/kralicky/ttr/pkg/game"
	"github.com/spf13/cobra"
)

// DirCmd represents the dir command
func BuildDirCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dir",
		Short: "Print the directory where game files are stored",
		Run: func(cmd *cobra.Command, args []string) {
			if dir, err := game.DataDir(); err != nil {
				fmt.Fprintln(os.Stderr, err.Error())
				os.Exit(1)
			} else {
				fmt.Println(dir)
			}
		},
	}
	return cmd
}
