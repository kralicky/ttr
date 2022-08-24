package config

import "github.com/spf13/viper"

const (
	accountsKey = "accounts"
)

func ListAccounts() []string {
	return viper.GetStringSlice(accountsKey)
}

func AddAccount(name string) {
	viper.Set(accountsKey, append(viper.GetStringSlice(accountsKey), name))
}

func AccountExists(name string) bool {
	accounts := viper.GetStringSlice(accountsKey)
	for _, account := range accounts {
		if account == name {
			return true
		}
	}
	return false
}

func DeleteAccount(name string) {
	accounts := viper.GetStringSlice(accountsKey)
	for i, account := range accounts {
		if account == name {
			accounts = append(accounts[:i], accounts[i+1:]...)
			break
		}
	}
	viper.Set(accountsKey, accounts)
}
