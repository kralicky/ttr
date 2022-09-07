package auth

import (
	"errors"
	"time"

	"github.com/pquerna/otp/totp"
	"github.com/zalando/go-keyring"
)

const (
	serviceName2fa = "ttr-cli-2fa"
)

func SetTwoFactorAuthSecret(accountName string, secret string) error {
	if _, err := keyring.Get(serviceName2fa, accountName); err == nil {
		return errors.New("2FA secret already exists for this account; delete it first")
	}
	return keyring.Set(serviceName2fa, accountName, secret)
}

func GetTwoFactorAuthSecret(accountName string) (string, error) {
	return keyring.Get(serviceName2fa, accountName)
}

func GenerateTwoFactorAuthCode(accountName string) (string, error) {
	secret, err := GetTwoFactorAuthSecret(accountName)
	if err != nil {
		return "", err
	}

	return totp.GenerateCode(secret, time.Now())
}

func HasTwoFactorAuthSecret(accountName string) bool {
	_, err := GetTwoFactorAuthSecret(accountName)
	return err == nil
}

func DeleteTwoFactorAuthSecret(accountName string) error {
	return keyring.Delete(serviceName2fa, accountName)
}
