package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

const (
	ConfigName = "settings"
	ConfigType = "json"
	EnvPrefix  = "GEMINI"
)

type Config struct {
	Project         string `mapstructure:"project"`
	Location        string `mapstructure:"location"`
	Model           string `mapstructure:"model"`
	MaxOutputTokens int    `mapstructure:"maxOutputTokens"`
	Temperature     float64 `mapstructure:"temperature"`
}

var DefaultConfig = Config{
	Model:           "gemini-1.5-pro",
	MaxOutputTokens: 8192,
	Temperature:     0.7,
	Location:        "us-central1",
}

func Init() error {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return err
	}
	configPath := filepath.Join(configDir, "gemini-cli")

	if err := os.MkdirAll(configPath, 0755); err != nil {
		return err
	}

	viper.AddConfigPath(configPath)
	viper.SetConfigName(ConfigName)
	viper.SetConfigType(ConfigType)
	viper.SetEnvPrefix(EnvPrefix)
	viper.AutomaticEnv()

	// Set defaults
	viper.SetDefault("model", DefaultConfig.Model)
	viper.SetDefault("maxOutputTokens", DefaultConfig.MaxOutputTokens)
	viper.SetDefault("temperature", DefaultConfig.Temperature)
	viper.SetDefault("location", DefaultConfig.Location)

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if desired
			// Create default config file? Maybe later.
		} else {
			return fmt.Errorf("error reading config file: %w", err)
		}
	}
	return nil
}

func Get() *Config {
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		// Should not happen if Init was called
		return &DefaultConfig
	}
	return &config
}

func Save() error {
	return viper.WriteConfig()
}

func Set(key string, value interface{}) {
	viper.Set(key, value)
}
