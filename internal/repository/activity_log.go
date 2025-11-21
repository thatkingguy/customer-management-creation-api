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
			"activityLogId", "description", "customerId", "createdAt", "updatedAt"
		) VALUES (
			$1, $2, $3, $4, $5
		)
	`

	now := time.Now()
	log.CreatedAt = now
	log.UpdatedAt = now

	_, err := tx.ExecContext(ctx, query,
		log.ActivityLogID,
		log.Description,
		log.CustomerID,
		log.CreatedAt,
		log.UpdatedAt,
	)

	return err
}
