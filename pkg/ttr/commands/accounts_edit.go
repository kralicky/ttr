/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package commands

import (
	"errors"

	"github.com/AlecAivazis/survey/v2"
	"github.com/kralicky/ttr/pkg/auth"
	"github.com/spf13/cobra"
)

// EditCmd represents the edit command
func BuildEditCmd() *cobra.Command {
	var editPassword bool
	cmd := &cobra.Command{
		Use:   "edit <username>",
		Short: "edit an existing account",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			qs := []*survey.Question{}
			if editPassword {
				qs = append(qs, &survey.Question{
					Name: "new-password",
					Prompt: &survey.Password{
						Message: "New Password:",
					},
				})
			}
			if len(qs) == 0 {
				return errors.New("nothing to edit (retry with --password)")
			}
			answers := struct {
				NewPassword string `survey:"new-password"`
			}{}
			if err := survey.Ask(qs, &answers); err != nil {
				return err
			}

			return auth.SetAccountPassword(args[0], answers.NewPassword)
		},
	}
	cmd.Flags().BoolVar(&editPassword, "reset-password", false, "reset password for the account")

	return cmd
}
