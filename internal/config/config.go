// Package config loads configuration from environment variables and an optional .env file.
package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

const (
	EnvProduction  = "production"
	EnvDevelopment = "development"
)

type Config struct {
	DB          DBConfig
	ServerPort  string
	JWTSecret   string
	JWTExpiry   time.Duration
	Environment string
}

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

// Load exits the process if a required variable is missing; a missing .env file is fine.
func Load() Config {
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		log.Fatalf("config: failed to load .env file: %v", err)
	}

	cfg := Config{
		DB: DBConfig{
			Host:     mustGetEnv("DB_HOST"),
			Port:     mustGetEnv("DB_PORT"),
			User:     mustGetEnv("DB_USER"),
			Password: mustGetEnv("DB_PASSWORD"),
			Name:     mustGetEnv("DB_NAME"),
			SSLMode:  mustGetEnv("DB_SSLMODE"),
		},
		ServerPort:  mustGetEnv("SERVER_PORT"),
		JWTSecret:   mustGetEnv("JWT_SECRET"),
		JWTExpiry:   mustGetDurationEnv("JWT_EXPIRY"),
		Environment: getEnvOrDefault("APP_ENV", EnvDevelopment),
	}

	return cfg
}

func getEnvOrDefault(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return fallback
}

func mustGetEnv(key string) string {
	val, ok := os.LookupEnv(key)
	if !ok || val == "" {
		log.Fatalf("config: required environment variable %q is not set", key)
	}
	return val
}

func mustGetDurationEnv(key string) time.Duration {
	val := mustGetEnv(key)
	d, err := time.ParseDuration(val)
	if err != nil {
		log.Fatalf("config: environment variable %q is not a valid duration: %v", key, err)
	}
	return d
}

// String is logged at startup, so it must never include secrets.
func (c Config) String() string {
	return fmt.Sprintf(
		"server_port=%s db_host=%s db_port=%s db_name=%s db_sslmode=%s jwt_expiry=%s environment=%s",
		c.ServerPort, c.DB.Host, c.DB.Port, c.DB.Name, c.DB.SSLMode, c.JWTExpiry, c.Environment,
	)
}

func (d DBConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		d.User, d.Password, d.Host, d.Port, d.Name, d.SSLMode,
	)
}
