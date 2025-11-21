package postgres

import (
	"context"
	"database/sql"
	"fmt"
)

// WithTransaction executes a function within a database transaction
func WithTransaction(ctx context.Context, db *sql.DB, fn func(*sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
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
			err = tx.Commit()
			if err != nil {
				err = fmt.Errorf("transaction commit error: %w", err)
			}
		}
	}()

	err = fn(tx)
	return err
}
