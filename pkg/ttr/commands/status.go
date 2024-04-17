package commands

import (
	"time"

	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/kralicky/ttr/pkg/api"
	"github.com/spf13/cobra"
)

func BuildStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show game status details",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()

			status, err := client.Status(cmd.Context())
			if err != nil {
				return err
			}
			if status.Open {
				cmd.Println(text.Colors{text.FgGreen}.Sprint("Game is open"))
			} else {
				cmd.Println(text.Colors{text.FgRed}.Sprint("Game is closed"))
			}
			if status.Banner != "" {
				cmd.Println(text.Colors{text.Bold, text.FgYellow}.Sprint(status.Banner))
			}
			if status.LastCookieIssuedAt > 0 {
				t := time.Unix(status.LastCookieIssuedAt, 0)
				cmd.Printf("Last cookie issued at: %s (%s ago)\n", t.Local().Format(time.RFC3339), time.Since(t).Truncate(time.Second))
			}
			if status.LastGameAuthAt > 0 {
				t := time.Unix(status.LastGameAuthAt, 0)
				cmd.Printf("Last game auth at:     %s (%s ago)\n", t.Local().Format(time.RFC3339), time.Since(t).Truncate(time.Second))
			}

			return nil
		},
	}
	return cmd
}
