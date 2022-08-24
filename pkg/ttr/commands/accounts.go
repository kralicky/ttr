/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package commands

import (
	"github.com/spf13/cobra"
)

// AccountsCmd represents the accounts command
func BuildAccountsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "accounts",
		Short: "Manage accounts and stored credentials",
	}
	return cmd
}
