/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package commands

import (
	"errors"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/kralicky/ttr/pkg/auth"
	"github.com/kralicky/ttr/pkg/config"
	"github.com/spf13/cobra"
	"github.com/zalando/go-keyring"
)

// ListCmd represents the list command
func BuildListCmd() *cobra.Command {
	var showSecrets bool
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List stored accounts",
		Run: func(cmd *cobra.Command, args []string) {
			w := table.NewWriter()
			w.SetStyle(table.StyleColoredDark)
			w.AppendHeader(table.Row{"ACCOUNT", "PASSWORD"})

			accounts := config.ListAccounts()
			for _, account := range accounts {
				password := "(secret)"
				if showSecrets {
					pw, err := auth.GetAccountPassword(account)
					if err != nil {
						if errors.Is(err, keyring.ErrNotFound) {
							password = "(not saved)"
						} else {
							password = "err: " + err.Error()
						}
					} else {
						password = pw
					}
				}
				w.AppendRow(table.Row{account, password})
			}

			cmd.Println(w.Render())
		},
	}
	cmd.Flags().BoolVar(&showSecrets, "show-secrets", false, "show stored passwords")
	return cmd
}
