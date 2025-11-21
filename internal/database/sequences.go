package database

import (
	"context"

	"github.com/sterling-retailcore-team/customer-management-creation-api/internal/logger"
)

// InitializeSequences creates sequences if they don't exist
func InitializeSequences(ctx context.Context) error {
	queries := []string{
		`CREATE SEQUENCE IF NOT EXISTS customer_number_seq START 1;`,
		`CREATE SEQUENCE IF NOT EXISTS customer_entity_id_seq START 1;`,
	}

	for _, query := range queries {
		if _, err := DB.ExecContext(ctx, query); err != nil {
			logger.Log.WithError(err).Errorf("Failed to create sequence: %s", query)
			return err
		}
	}

	logger.Log.Info("Sequences initialized successfully")
	return nil
}

