// Package config loads application configuration from environment variables,
// with optional .env file support for local development via godotenv.
package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all required application configuration.
type Config struct {
	DB         DBConfig
	ServerPort string
	JWTSecret  string
	JWTExpiry  time.Duration
}

// DBConfig holds database connection settings.
type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

// Load reads configuration from environment variables, loading a .env file
// first if one is present (missing .env is not an error — e.g. in prod where
// vars are set directly). It exits the process with a clear message if any
// required variable is missing.
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
		ServerPort: mustGetEnv("SERVER_PORT"),
		JWTSecret:  mustGetEnv("JWT_SECRET"),
		JWTExpiry:  mustGetDurationEnv("JWT_EXPIRY"),
	}

	return cfg
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

// String returns a summary of the loaded config safe for logging — it never
// includes secret values (JWT secret, DB password).
func (c Config) String() string {
	return fmt.Sprintf(
		"server_port=%s db_host=%s db_port=%s db_name=%s db_sslmode=%s jwt_expiry=%s",
		c.ServerPort, c.DB.Host, c.DB.Port, c.DB.Name, c.DB.SSLMode, c.JWTExpiry,
	)
}

// DSN returns the PostgreSQL connection string for this config, suitable for
// use with the golang-migrate CLI or a database driver.
func (d DBConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		d.User, d.Password, d.Host, d.Port, d.Name, d.SSLMode,
	)
}
