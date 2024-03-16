package commands

import (
	"github.com/kralicky/ttr/pkg/multi"
	"github.com/spf13/cobra"
)

func BuildMultitoonCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "multitoon",
		Short: "Utility for duplicating inputs to multiple windows",
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr, err := multi.NewManager()
			if err != nil {
				return err
			}
			mgr.RunInputWindow()
			<-cmd.Context().Done()
			mgr.Dispose()
			return nil
		},
	}
	return cmd
}
