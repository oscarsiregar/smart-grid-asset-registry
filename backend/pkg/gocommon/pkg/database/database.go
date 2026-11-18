package database

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

const (
	defaultMaxIdleConns    = 10
	defaultMaxOpenConns    = 100
	defaultConnMaxLifetime = 60 // minutes
	slowQueryThreshold     = 200 * time.Millisecond
)

// Connect opens a GORM connection to PostgreSQL and configures the pool.
// Zero pool values fall back to the defaults above.
func Connect(cfg *DBConfig, log *logrus.Logger, isProd bool) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=UTC",
		cfg.Host, cfg.User, cfg.Password, cfg.DBName, cfg.Port,
	)

	gormConfig := &gorm.Config{
		SkipDefaultTransaction: true,
		TranslateError:         true,
	}
	if log != nil {
		level := gormlogger.Info
		if isProd {
			level = gormlogger.Warn
		}
		gormConfig.Logger = gormlogger.New(log, gormlogger.Config{
			SlowThreshold:             slowQueryThreshold,
			LogLevel:                  level,
			IgnoreRecordNotFoundError: true,
		})
	}

	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql db: %w", err)
	}

	sqlDB.SetMaxIdleConns(valueOrDefault(cfg.MaxIdleConns, defaultMaxIdleConns))
	sqlDB.SetMaxOpenConns(valueOrDefault(cfg.MaxOpenConns, defaultMaxOpenConns))
	sqlDB.SetConnMaxLifetime(time.Duration(valueOrDefault(cfg.ConnMaxLifetime, defaultConnMaxLifetime)) * time.Minute)

	return db, nil
}

func valueOrDefault(value, fallback int) int {
	if value <= 0 {
		return fallback
	}
	return value
}
