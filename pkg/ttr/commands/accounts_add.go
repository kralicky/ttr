package commands

import (
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	"github.com/kralicky/ttr/pkg/auth"
	"github.com/kralicky/ttr/pkg/config"
	"github.com/spf13/cobra"
)

// AddCmd represents the add command
func BuildAddCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Args:  cobra.NoArgs,
		Short: "Add new account credentials",
		RunE: func(cmd *cobra.Command, args []string) error {
			qs := []*survey.Question{
				{
					Name: "name",
					Prompt: &survey.Input{
						Message: "Username or email:",
					},
					Validate: survey.Required,
				},
				{
					Name: "password",
					Prompt: &survey.Password{
						Message: "Password:",
					},
				},
			}
			answers := struct {
				Name     string `survey:"name"`
				Password string `survey:"password"`
			}{}
			err := survey.Ask(qs, &answers)
			if err != nil {
				return err
			}

			exists := config.AccountExists(answers.Name)
			if exists {
				return fmt.Errorf("account %s already exists", answers.Name)
			}

			config.AddAccount(answers.Name)
			if err := config.Save(); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

			if len(answers.Password) > 0 {
				if err := auth.SetAccountPassword(answers.Name, answers.Password); err != nil {
					return fmt.Errorf("failed to store credentials: %w", err)
				}
			} else {
				// clear an existing password if any were left over, ignore any errors
				auth.DeleteAccountPassword(answers.Name)
				cmd.Println("Password will not be saved for this account (you will be prompted for it every time)")
			}

			return nil
		},
	}
	return cmd
}
