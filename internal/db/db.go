// Package db opens the shared *gorm.DB connection used across modules.
package db

import (
	"fmt"
	"time"

	"github.com/AbolfazlTayyar/gogym/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	maxOpenConns    = 25
	maxIdleConns    = 25
	connMaxLifetime = 5 * time.Minute
)

// New never calls AutoMigrate: the SQL migrations are the only source of truth for the schema.
func New(cfg config.Config) (*gorm.DB, error) {
	gormDB, err := gorm.Open(postgres.Open(cfg.DB.DSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel(cfg.Environment)),
		// Lets repositories match portable errors like gorm.ErrDuplicatedKey.
		TranslateError: true,
	})
	if err != nil {
		return nil, fmt.Errorf("db: failed to open connection: %w", err)
	}

	sqlDB, err := gormDB.DB()
	if err != nil {
		return nil, fmt.Errorf("db: failed to get underlying sql.DB: %w", err)
	}

	sqlDB.SetMaxOpenConns(maxOpenConns)
	sqlDB.SetMaxIdleConns(maxIdleConns)
	sqlDB.SetConnMaxLifetime(connMaxLifetime)

	return gormDB, nil
}

func logLevel(environment string) logger.LogLevel {
	if environment == config.EnvProduction {
		return logger.Warn
	}
	return logger.Info
}
