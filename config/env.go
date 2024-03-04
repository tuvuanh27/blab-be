package config

import "github.com/spf13/viper"

type Config struct {
	Port     string `env:"PORT"`
	RedisUrl string `env:"REDIS_URL"`
}

func LoadEnv() (cfg Config, err error) {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()
	viper.SetConfigType("env")

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
