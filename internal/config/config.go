package config

import (
	"os"
	"strconv"
	"time"

	"github.com/keykibatyr/triad-chat/internal/models"
)

type ServerConfig struct {
	Port int
}

type JWTConfig struct {
	AccessTokenSecret  []byte
	RefreshTokenSecret []byte
	AccessTokenTTL     time.Duration
	RefreshTokenTTL    time.Duration
	Issuer             string
}

type Config struct {
	Database models.PostgresConfig
	Server   ServerConfig
	JWT 	JWTConfig
}

func Load() *Config {
	return &Config{
		Database: models.PostgresConfig{
			Host:     getEnv("DB_HOST", "db"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "keykibatyr"),
			Password: getEnv("DB_PASSWORD", "Alisher0505"),
			Database: getEnv("DB_DATABASE", "triadchat"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},

		Server: ServerConfig{
			Port: getEnvInt("SERVER_PORT", 8080),
		},

		JWT: JWTConfig{
			AccessTokenSecret: []byte(getEnv("JWT_ACCESS_SECRET", "something")),
			RefreshTokenSecret: []byte(getEnv("JWT_REFRESH_SECRET", "something")),
			AccessTokenTTL: 15 * time.Minute,
			RefreshTokenTTL: 7 * 24 * time.Hour,
			Issuer: "triad-chat",
		},
	}
}

func getEnv(key string, defaultValue string) string {
	envValue := os.Getenv(key)
	if envValue != "" {
		return envValue
	}

	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	envValue := os.Getenv(key)
	if envValue != "" {
		intVal, err := strconv.Atoi(envValue)
		if err != nil {
			return defaultValue
		}
		return intVal
	}

	return defaultValue
}
