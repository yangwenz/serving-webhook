package utils

import (
	"github.com/spf13/viper"
)

// Config stores all configuration of the application.
// The values are read by viper from a config file or environment variable.
type Config struct {
	HTTPServerAddress  string `mapstructure:"HTTP_SERVER_ADDRESS"`
	RedisAddress       string `mapstructure:"REDIS_ADDRESS"`
	AWSBucket          string `mapstructure:"AWS_BUCKET"`
	AWSRegion          string `mapstructure:"AWS_REGION"`
	AWSS3UseAccelerate bool   `mapstructure:"AWS_S3_USE_ACCELERATE"`
}

// LoadConfig reads configuration from file or environment variables.
func LoadConfig(path string) (config Config, err error) {
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
