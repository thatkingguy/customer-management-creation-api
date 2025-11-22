package repository

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/sterling-retailcore-team/customer-management-creation-api/internal/models"
)

type ReferenceRepository interface {
	Create(ctx context.Context, tx *sql.Tx, reference *models.Reference) error
	CreateBatch(ctx context.Context, tx *sql.Tx, references []*models.Reference) error
}

type referenceRepository struct {
	db *sqlx.DB
}

func NewReferenceRepository(db *sqlx.DB) ReferenceRepository {
	return &referenceRepository{db: db}
}

func (r *referenceRepository) Create(ctx context.Context, tx *sql.Tx, reference *models.Reference) error {
	query := `
		INSERT INTO reference (
			"referenceId", "customerId", "name", "phoneNumber", "email",
			"address", "relationship", "createdAt", "updatedAt"
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9
		)
	`

	_, err := tx.ExecContext(ctx, query,
		reference.ReferenceID,
		reference.CustomerID,
		reference.Name,
		reference.PhoneNumber,
		reference.Email,
		reference.Address,
		reference.Relationship,
		reference.CreatedAt,
		reference.UpdatedAt,
	)

	return err
}

func (r *referenceRepository) CreateBatch(ctx context.Context, tx *sql.Tx, references []*models.Reference) error {
	if len(references) == 0 {
		return nil
	}

	for _, reference := range references {
		if err := r.Create(ctx, tx, reference); err != nil {
			return err
		}
	}

	return nil
}

