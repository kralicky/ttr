package config

import (
	"os"

	"github.com/spf13/viper"
)

func Load(filename string) error {
	viper.SetConfigFile(filename)
	setDefaults()

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		if err := viper.SafeWriteConfigAs(filename); err != nil {
			return err
		}
	}

	return viper.ReadInConfig()
}

func Save() error {
	return viper.WriteConfig()
}

func setDefaults() {
	viper.SetDefault(accountsKey, []string{})
}
