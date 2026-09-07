package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	Auth     AuthConfig
}

type ServerConfig struct {
	Port    string
	GinMode string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}
type JWTConfig struct {
	Secret              string
	ExpiresIn           time.Duration
	RefreshTokenExpires time.Duration
}

// AuthConfig is the cost of a password hash and the shape of the lockout: how
// many failures in a row close an account, and for how long.
type AuthConfig struct {
	BcryptCost             int
	MaxFailedLoginAttempts int
	LockoutDuration        time.Duration
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	jwtExpiresIn, _ := time.ParseDuration(getEnv("JWT_EXPIRES_IN", "24h"))
	refreshTokenExpires, _ := time.ParseDuration(getEnv("REFRESH_TOKEN_EXPIRES_IN", "720h"))
	lockoutDuration, _ := time.ParseDuration(getEnv("LOCKOUT_DURATION", "15m"))
	return &Config{
		Server: ServerConfig{
			Port:    getEnv("PORT", "8090"),
			GinMode: getEnv("GIN_MODE", "debug"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "password"),
			Name:     getEnv("DB_NAME", "go_auth"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		JWT: JWTConfig{
			Secret:              getEnv("JWT_SECRET", "you_jwt_secret_key"),
			ExpiresIn:           jwtExpiresIn,
			RefreshTokenExpires: refreshTokenExpires,
		},
		Auth: AuthConfig{
			BcryptCost:             getEnvInt("BCRYPT_COST", 12),
			MaxFailedLoginAttempts: getEnvInt("MAX_FAILED_LOGIN_ATTEMPTS", 5),
			LockoutDuration:        lockoutDuration,
		},
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	value, err := strconv.Atoi(os.Getenv(key))
	if err != nil {
		return defaultValue
	}
	return value
}
