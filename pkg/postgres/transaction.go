package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/sterling-retailcore-team/customer-management-creation-api/internal/logger"
)

// WithTransaction executes a function within a database transaction
func WithTransaction(ctx context.Context, db *sql.DB, fn func(*sql.Tx) error) error {
	beginStart := time.Now()
	tx, err := db.BeginTx(ctx, nil)
	beginDuration := time.Since(beginStart)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	if beginDuration > 10*time.Millisecond {
		logger.Log.WithField("duration_ms", beginDuration.Milliseconds()).Debug("WithTransaction: BeginTx took longer than expected")
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		} else if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				err = fmt.Errorf("transaction rollback error: %v, original error: %w", rbErr, err)
			}
		} else {
			// Measure only the commit operation itself
			commitStart := time.Now()
			err = tx.Commit()
			commitDuration := time.Since(commitStart)
			if err != nil {
				err = fmt.Errorf("transaction commit error: %w", err)
			} else if commitDuration > 10*time.Millisecond {
				logger.Log.WithField("duration_ms", commitDuration.Milliseconds()).Debug("WithTransaction: Commit took longer than expected")
			}
		}
	}()

	err = fn(tx)
	return err
}
