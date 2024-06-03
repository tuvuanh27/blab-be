package config

import "github.com/spf13/viper"

type Config struct {
	Port     string `mapstructure:"PORT"`
	RedisUrl string `mapstructure:"REDIS_URL"`
	Rpc      string `mapstructure:"RPC"`
}

func LoadEnv() (cfg Config, err error) {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()
	viper.SetConfigType("")

	err = viper.ReadInConfig()
	if err != nil {
		panic(err)
	}

	err = viper.Unmarshal(&cfg)
	if err != nil {
		panic(err)
	}

	return
}

var ConfigEnv, _ = LoadEnv()
