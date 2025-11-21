package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/sterling-retailcore-team/customer-management-creation-api/internal/config"
	"github.com/sterling-retailcore-team/customer-management-creation-api/internal/logger"
)

var DB *sqlx.DB

// Connect establishes a database connection with retry logic
func Connect(cfg *config.DatabaseConfig) error {
	var err error
	maxRetries := 5
	retryDelay := 2 * time.Second

	for i := 0; i < maxRetries; i++ {
		logger.Log.Debugf("Attempting database connection (attempt %d/%d)", i+1, maxRetries)
		DB, err = sqlx.Connect("postgres", cfg.ConnectionString())
		if err == nil {
			// Configure connection pool
			DB.SetMaxOpenConns(cfg.MaxOpenConns)
			DB.SetMaxIdleConns(cfg.MaxIdleConns)
			DB.SetConnMaxLifetime(cfg.ConnMaxLifetime)
			DB.SetConnMaxIdleTime(10 * time.Minute)
			logger.Log.WithFields(map[string]interface{}{
				"max_open_conns":    cfg.MaxOpenConns,
				"max_idle_conns":    cfg.MaxIdleConns,
				"conn_max_lifetime": cfg.ConnMaxLifetime,
			}).Debug("Database connection pool configured")

			// Test connection
			logger.Log.Debug("Pinging database to verify connection")
			if err = DB.Ping(); err == nil {
				logger.Log.Info("Database connection established successfully")
				return nil
			}
			logger.Log.WithError(err).Error("Database ping failed")
		} else {
			logger.Log.WithError(err).Errorf("Failed to connect to database (attempt %d/%d)", i+1, maxRetries)
		}

		if i < maxRetries-1 {
			logger.Log.Warnf("Failed to connect to database (attempt %d/%d): %v. Retrying...", i+1, maxRetries, err)
			time.Sleep(retryDelay)
			retryDelay *= 2 // Exponential backoff
		}
	}

	return fmt.Errorf("failed to connect to database after %d attempts: %w", maxRetries, err)
}

// Close closes the database connection gracefully
func Close() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}

// HealthCheck checks if the database connection is healthy
func HealthCheck(ctx context.Context) error {
	if DB == nil {
		return fmt.Errorf("database connection is nil")
	}
	return DB.PingContext(ctx)
}

// GetDB returns the database connection
func GetDB() *sqlx.DB {
	return DB
}

// BeginTx starts a new transaction
func BeginTx(ctx context.Context) (*sqlx.Tx, error) {
	return DB.BeginTxx(ctx, nil)
}
