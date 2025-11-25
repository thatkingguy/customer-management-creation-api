package repository

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/sterling-retailcore-team/customer-management-creation-api/internal/models"
)

type OtherAccountRepository interface {
	Create(ctx context.Context, tx *sql.Tx, account *models.OtherAccount) error
	CreateBatch(ctx context.Context, tx *sql.Tx, accounts []*models.OtherAccount) error
}

type otherAccountRepository struct {
	db *sqlx.DB
}

func NewOtherAccountRepository(db *sqlx.DB) OtherAccountRepository {
	return &otherAccountRepository{db: db}
}

func (r *otherAccountRepository) Create(ctx context.Context, tx *sql.Tx, account *models.OtherAccount) error {
	query := `
		INSERT INTO other_account (
			"otherAccountId", "customerId", "customerProfileId", "bankName",
			"accountNumber", "accountType", "createdAt", "updatedAt"
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8
		)
	`

	_, err := tx.ExecContext(ctx, query,
		account.OtherAccountID,
		account.CustomerID,
		account.CustomerProfileID,
		account.BankName,
		account.AccountNumber,
		account.AccountType,
		account.CreatedAt,
		account.UpdatedAt,
	)

	return err
}

func (r *otherAccountRepository) CreateBatch(ctx context.Context, tx *sql.Tx, accounts []*models.OtherAccount) error {
	if len(accounts) == 0 {
		return nil
	}

	for _, account := range accounts {
		if err := r.Create(ctx, tx, account); err != nil {
			return err
		}
	}

	return nil
}




