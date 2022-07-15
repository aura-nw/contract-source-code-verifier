package util

import (
	"smart-contract-verify/model"

	"github.com/spf13/viper"
)

func LoadConfig(path string) (config model.Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&config)
	return
}
