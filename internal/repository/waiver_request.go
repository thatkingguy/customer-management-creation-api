package repository

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type WaiverRequestRepository interface {
	UpdateCustomerID(ctx context.Context, tx *sql.Tx, requestID uuid.UUID, customerID uuid.UUID) error
}

type waiverRequestRepository struct {
	db *sqlx.DB
}

func NewWaiverRequestRepository(db *sqlx.DB) WaiverRequestRepository {
	return &waiverRequestRepository{db: db}
}

func (r *waiverRequestRepository) UpdateCustomerID(ctx context.Context, tx *sql.Tx, requestID uuid.UUID, customerID uuid.UUID) error {
	query := `UPDATE waiver_request SET "customerId" = $2, "updatedAt" = NOW() WHERE "requestId" = $1`
	_, err := tx.ExecContext(ctx, query, requestID, customerID)
	return err
}

