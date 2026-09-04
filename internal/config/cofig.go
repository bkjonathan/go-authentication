package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Server ServerConfig
}

type ServerConfig struct {
	Port    string
	GinMode string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	return &Config{
		Server: ServerConfig{
			Port:    getEnv("PORT", "8090"),
			GinMode: getEnv("GIN_MODE", "debug"),
		},
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
