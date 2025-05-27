package util

import (
	"github.com/spf13/viper"
)

type Config struct {
	DbDriver           string `mapstructure:"db_driver"`
	DbConnectionString string `mapstructure:"db_connection_string"`
	WebServerAddress   string `mapstructure:"web_server_address"`
}

func LoadConfig(configPath string) (Config, error) {
	var cfg Config
	viper.AddConfigPath(configPath)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AutomaticEnv()
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		return cfg, err
	}

	viper.Unmarshal(&cfg)

	return cfg, nil

}
