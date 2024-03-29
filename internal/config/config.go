package config

import (
	"github.com/spf13/viper"
	"go.uber.org/fx"
)

type DatabaseConfig struct {
	Host       string
	Port       int
	Name       string
	User       string
	Password   string
	DisableSSL bool
}

type Result struct {
	fx.Out

	DatabaseConfig DatabaseConfig
}

func newConfig() Result {
	viper.SetDefault("env", "local")
	viper.AddConfigPath(".")
	viper.SetConfigName(viper.GetString("env"))

	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}

	return Result{
		DatabaseConfig: DatabaseConfig{
			Host:       viper.GetString("database.host"),
			Port:       viper.GetInt("database.port"),
			Name:       viper.GetString("database.name"),
			User:       viper.GetString("database.user"),
			Password:   viper.GetString("database.password"),
			DisableSSL: viper.GetBool("database.disable_ssl"),
		},
	}
}

var Module = fx.Options(
	fx.Provide(
		newConfig,
	),
)
