package util

import (
	"os"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	DbDriver                 string        `mapstructure:"db_driver"`
	DbConnectionString       string        `mapstructure:"db_connection_string"`
	WebServerAddress         string        `mapstructure:"web_server_address"`
	JwtSecretKey             string        `mapstructure:"jwt_secret_key"`
	JwtTokenDuration         time.Duration `mapstructure:"jwt_token_duration"`
	JWtRefreshTokenDuration  time.Duration `mapstructure:"jwt_refresh_token_duration"`
	GrpcServerAddress        string        `mapstructure:"grpc_server_address"`
}

func LoadConfig(configPath string) (Config, error) {
	var cfg Config
	confingFileName := "config"
	if path, exists := os.LookupEnv("CONFIG_FILE_NAME"); exists {
		confingFileName = path
	}
	viper.AddConfigPath(configPath)
	viper.SetConfigName(confingFileName)
	viper.SetConfigType("yaml")
	viper.AutomaticEnv()
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		return cfg, err
	}

	viper.Unmarshal(&cfg)

	return cfg, nil

}
