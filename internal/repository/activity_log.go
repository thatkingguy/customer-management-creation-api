package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/sterling-retailcore-team/customer-management-creation-api/internal/models"
)

type ActivityLogRepository interface {
	Create(ctx context.Context, tx *sql.Tx, log *models.ActivityLog) error
}

type activityLogRepository struct {
	db *sqlx.DB
}

func NewActivityLogRepository(db *sqlx.DB) ActivityLogRepository {
	return &activityLogRepository{db: db}
}

func (r *activityLogRepository) Create(ctx context.Context, tx *sql.Tx, log *models.ActivityLog) error {
	query := `
		INSERT INTO activity_log (
			"activityLogId", "customerId", "requestId", "description", "reason", "createdAt", "updatedAt"
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7
		)
	`

	now := time.Now()
	log.CreatedAt = now
	log.UpdatedAt = now

	var reasonJSON interface{}
	if log.Reason != nil {
		reasonJSON = log.Reason
	}

	args := []interface{}{
		log.ActivityLogID,
		log.CustomerID,
		log.RequestID,
		log.Description,
		reasonJSON,
		log.CreatedAt,
		log.UpdatedAt,
	}

	// Use transaction if provided, otherwise use DB connection directly
	if tx != nil {
		_, err := tx.ExecContext(ctx, query, args...)
		return err
	}

	_, err := r.db.ExecContext(ctx, query, args...)
	return err
}
