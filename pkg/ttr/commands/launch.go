package commands

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/kralicky/ttr/pkg/api"
	"github.com/kralicky/ttr/pkg/auth"
	"github.com/kralicky/ttr/pkg/config"
	"github.com/kralicky/ttr/pkg/game"
	"github.com/spf13/cobra"
)

// LaunchCmd represents the launch command
func BuildLaunchCmd() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "launch",
		Short: "Launch the TTR engine",
		RunE: func(cmd *cobra.Command, args []string) error {
			// check for updates in the background
			client := api.NewClient()
			doneUpdating := make(chan error, 1)
			go func() {
				defer close(doneUpdating)
				if err := game.SyncGameData(cmd.Context(), client); err != nil {
					doneUpdating <- err
				}
			}()

			// prompt for account
			accounts := config.ListAccounts()
			if len(accounts) == 0 {
				return fmt.Errorf("no accounts found, run `ttr accounts add` to add one.")
			}

			var selected []string
			if err := survey.AskOne(&survey.MultiSelect{
				Message: "Select accounts:",
				Options: accounts,
			}, &selected); err != nil {
				return err
			}

			var wg sync.WaitGroup

			for _, account := range selected {
				account := account

				pw, err := auth.GetAccountPasswordOrPrompt(account)
				if err != nil {
					return err
				}

				resp, err := client.Login(cmd.Context(), account, pw)
				if err != nil {
					return err
				}
				switch resp.Success {
				case api.SuccessTrue:

				case api.SuccessFalse:
					return fmt.Errorf("login failed: %s", resp.Message)
				case api.SuccessPartial:
					return fmt.Errorf("two-factor authentication not supported yet")
				case api.SuccessDelayed:
					for {
						time.Sleep(1 * time.Second)
						retryResp, err := client.RetryDelayedLogin(cmd.Context(), resp.QueueToken)
						if err == nil {
							if retryResp.Success == api.SuccessTrue {
								resp = retryResp
								break
							}
						}
					}

				}

				// wait for updates to finish
				if err := <-doneUpdating; err != nil {
					return err
				}

				wg.Add(1)
				go func() {
					defer wg.Done()
					fmt.Printf("Running: %s\n", account)
					if err := game.LaunchProcess(cmd.Context(), resp.LoginSuccessPayload); err != nil {
						fmt.Fprintln(os.Stderr, err)
					}
					fmt.Printf("Exited: %s\n", account)
				}()
			}
			wg.Wait()
			return nil
		},
	}
	return cmd
}
