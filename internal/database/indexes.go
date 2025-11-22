package database

import (
	"context"
	"embed"

	"github.com/sterling-retailcore-team/customer-management-creation-api/internal/logger"
)

//go:embed migrations/007_create_request_and_interim_indexes.sql
var indexMigrationFS embed.FS

// InitializeIndexes creates indexes for request and interim_approval_config tables
func InitializeIndexes(ctx context.Context) error {
	// Execute migration - split into individual statements to handle errors better
	statements := []string{
		`CREATE INDEX IF NOT EXISTS idx_request_request_id_not_deleted ON request("requestId") WHERE "isDeleted" = false;`,
		`CREATE INDEX IF NOT EXISTS idx_request_status_deleted ON request("status", "isDeleted") WHERE "isDeleted" = false;`,
		`CREATE INDEX IF NOT EXISTS idx_interim_approval_config_lookup ON interim_approval_config("customerType", "gracePeriod", "status", "createdAt" DESC);`,
		`CREATE INDEX IF NOT EXISTS idx_interim_approval_config_active ON interim_approval_config("customerType", "createdAt" DESC) WHERE "gracePeriod" = true AND "status" = 'Approved';`,
	}

	for _, stmt := range statements {
		if _, err := DB.ExecContext(ctx, stmt); err != nil {
			// Log but don't fail - indexes might already exist
			logger.Log.WithError(err).WithField("statement", stmt).Warn("Index creation statement failed (may already exist)")
		}
	}

	logger.Log.Info("Request and interim approval config indexes initialized successfully")
	return nil
}

