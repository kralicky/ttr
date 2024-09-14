package commands

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/kralicky/ttr/pkg/api"
	"github.com/kralicky/ttr/pkg/auth"
	"github.com/kralicky/ttr/pkg/config"
	"github.com/kralicky/ttr/pkg/game"
	"github.com/spf13/cobra"
)

// LaunchCmd represents the launch command
func BuildLaunchCmd() *cobra.Command {
	var skipUpdateCheck bool
	cmd := &cobra.Command{
		Use:   "launch",
		Short: "Launch the TTR engine",
		PreRun: func(cmd *cobra.Command, args []string) {
			go game.RunGLFW()
		},
		PostRun: func(cmd *cobra.Command, args []string) {
			game.ShutdownGLFW()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			// check for updates in the background
			client := api.NewClient()

			status, err := client.Status(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get game status: %w", err)
			}
			banner := status.Banner
			if !status.Open {
				if banner == "" {
					banner = "(no details given)"
				}
				cmd.Printf(text.Colors{text.FgRed}.Sprintf("Game may be closed: %s\n", banner))
			} else if banner != "" {
				cmd.Printf(text.Colors{text.Bold, text.FgYellow}.Sprintf("%s\n\n", banner))
			}
			doneUpdating := make(chan error, 1)
			if skipUpdateCheck {
				close(doneUpdating)
			} else {
				go func() {
					defer close(doneUpdating)
					if err := game.SyncGameData(cmd.Context(), client); err != nil {
						doneUpdating <- err
					}
				}()
			}

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
					return fmt.Errorf("login failed: %w", err)
				}
			RESPONSE:
				for {
					switch resp.Success {
					case api.SuccessTrue:
						break RESPONSE
					case api.SuccessFalse:
						return fmt.Errorf("login failed: %s", resp.Message)
					case api.SuccessPartial:
						var code string
						if auth.HasTwoFactorAuthSecret(account) {
							var err error
							fmt.Printf("Generating two-factor authentication code for %s...\n", account)
							code, err = auth.GenerateTwoFactorAuthCode(account)
							if err != nil {
								fmt.Fprintf(os.Stderr, "error generating two-factor authentication code: %s", err)
								return err
							}
						} else {
							if err := survey.AskOne(&survey.Password{
								Message: "Enter a two-factor authentication code for " + account + ":",
							}, &code); err != nil {
								return err
							}
						}
						resp, err = client.CompleteTwoFactorAuth(cmd.Context(), resp.ResponseToken, code)
						if err != nil {
							return fmt.Errorf("two-factor authentication failed: %w", err)
						}
					case api.SuccessDelayed:
						if resp.ETA > 0 {
							time.Sleep(1 * time.Second)
						}
						resp, err = client.RetryDelayedLogin(cmd.Context(), resp.QueueToken)
						if err != nil {
							return fmt.Errorf("failed to retry delayed login: %w", err)
						}
					}
				}

				// wait for updates to finish
				if err := <-doneUpdating; err != nil {
					return fmt.Errorf("update failed: %w", err)
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

	cmd.Flags().BoolVar(&skipUpdateCheck, "skip-update-check", false, "Skip checking for updates")
	return cmd
}
