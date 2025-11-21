package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/sterling-retailcore-team/customer-management-creation-api/internal/models"
)

type CustomerRepository interface {
	Create(ctx context.Context, tx *sql.Tx, customer *models.Customer) error
}

type customerRepository struct {
	db *sqlx.DB
}

func NewCustomerRepository(db *sqlx.DB) CustomerRepository {
	return &customerRepository{db: db}
}

func (r *customerRepository) Create(ctx context.Context, tx *sql.Tx, customer *models.Customer) error {
	query := `
		INSERT INTO customer (
			"customerId", "customerType", "status", "monitoring", "approvalStatus",
			"tenantId", "initiator", "initiatorId", "approver", "approverId",
			"branch", "approverBranch", "department", "categoryOfCustomer",
			"customerSubType", "relatedEntity", "createdAt", "updatedAt"
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18
		)
	`

	now := time.Now()
	customer.CreatedAt = now
	customer.UpdatedAt = now

	_, err := tx.ExecContext(ctx, query,
		customer.CustomerID,
		customer.CustomerType,
		customer.Status,
		customer.Monitoring,
		customer.ApprovalStatus,
		customer.TenantID,
		customer.Initiator,
		customer.InitiatorID,
		customer.Approver,
		customer.ApproverID,
		customer.Branch,
		customer.ApproverBranch,
		customer.Department,
		customer.CategoryOfCustomer,
		customer.CustomerSubType,
		customer.RelatedEntity,
		customer.CreatedAt,
		customer.UpdatedAt,
	)

	return err
}

