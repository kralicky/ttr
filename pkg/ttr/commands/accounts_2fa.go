package commands

import (
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	"github.com/kralicky/ttr/pkg/auth"
	"github.com/spf13/cobra"
)

func BuildTwoFactorAuthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "2fa",
		Short: "two factor authentication commands",
	}

	cmd.AddCommand(BuildSetupTwoFactorAuthCmd())
	cmd.AddCommand(BuildForgetTwoFactorAuthCmd())
	cmd.AddCommand(BuildGenerateCodeCmd())

	return cmd
}

func BuildSetupTwoFactorAuthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "setup <username>",
		Short: "set up two-factor authentication for an account",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var confirm bool
			survey.AskOne(&survey.Confirm{
				Message: `
This will set up automatic two-factor authentication for your account without
requiring you to enter a code from an authenticator app. This is a beta 
feature and may have bugs! It is STRONGLY recommended either to use an 
authenticator app in addition to this feature (by scanning the QR code) or to
keep backups of your recovery codes somewhere safe.

If your login keyring is lost or destroyed, and you lose your recovery codes, 
you will lose access to your account and will need to contact support to regain 
access to it.

Do you understand the potential risks, and wish to continue?`[1:],
			}, &confirm)
			if !confirm {
				fmt.Println("Aborting setup.")
				return nil
			}

			var secret string
			if err := survey.AskOne(&survey.Password{
				Message: "Enter the Two-Step Login code given by the TTR account page:",
				Help:    "This can be found by clicking \"QR code not working?\" under the QR code.",
			}, &secret); err != nil {
				return err
			}

			err := auth.SetTwoFactorAuthSecret(args[0], secret)
			if err != nil {
				return err
			}

			var generateTestCode bool
			if err := survey.AskOne(&survey.Confirm{
				Message: "Secret saved. Generate a test code?",
				Default: true,
			}, &generateTestCode); err != nil {
				return err
			}

			if generateTestCode {
				code, err := auth.GenerateTwoFactorAuthCode(args[0])
				if err != nil {
					return fmt.Errorf("failed to generate a TOTP code: %w", err)
				}

				fmt.Println(fmt.Sprintf("Your code is: %s", code))
			}
			return nil
		},
	}

	return cmd
}

func BuildForgetTwoFactorAuthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "forget <username>",
		Short: "forget two-factor authentication for an account",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var confirm bool
			if err := survey.AskOne(&survey.Confirm{
				Message: "Are you sure you want to delete your two-factor authentication secret for " + args[0] + "?",
			}, &confirm); err != nil {
				return err
			}
			if !confirm {
				fmt.Println("Aborting.")
				return nil
			}
			if err := survey.AskOne(&survey.Confirm{
				Message: "Please note that you MUST disable two-factor auth on the TTR account page before running this command. Continue?",
			}, &confirm); err != nil {
				return err
			}
			if !confirm {
				fmt.Println("Aborting.")
				return nil
			}

			err := auth.DeleteTwoFactorAuthSecret(args[0])
			if err != nil {
				return err
			}

			fmt.Println("Two-factor auth secret deleted.")
			return nil
		},
	}

	return cmd
}

func BuildGenerateCodeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate <username>",
		Short: "generate a two-factor authentication code",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			code, err := auth.GenerateTwoFactorAuthCode(args[0])
			if err != nil {
				return fmt.Errorf("failed to generate a two-factor auth code: %w", err)
			}

			fmt.Println(fmt.Sprintf("Your code is: %s", code))
			return nil
		},
	}

	return cmd
}
