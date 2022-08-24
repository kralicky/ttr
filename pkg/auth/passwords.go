package auth

import (
	"errors"

	"github.com/AlecAivazis/survey/v2"
	"github.com/zalando/go-keyring"
)

const (
	serviceName = "ttr-cli"
)

func GetAccountPassword(user string) (string, error) {
	return keyring.Get(serviceName, user)
}

func GetAccountPasswordOrPrompt(user string) (string, error) {
	pw, err := keyring.Get(serviceName, user)
	if err == nil {
		return pw, nil
	}
	if errors.Is(err, keyring.ErrNotFound) {
		var password string
		if err := survey.AskOne(&survey.Password{
			Message: "Password:",
		}, &password); err != nil {
			return "", err
		}
		return password, nil
	}
	return "", err
}

func SetAccountPassword(user, password string) error {
	return keyring.Set(serviceName, user, password)
}

func DeleteAccountPassword(user string) error {
	err := keyring.Delete(serviceName, user)
	if errors.Is(err, keyring.ErrNotFound) {
		return nil
	}
	return err
}
