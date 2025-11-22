package service

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/sterling-retailcore-team/customer-management-creation-api/internal/repository"
)

type IDGeneratorService interface {
	GenerateUUID() uuid.UUID
	GenerateCustomerNumber(ctx context.Context, tx *sql.Tx, profileRepo repository.CustomerProfileRepository) (string, error)
	GenerateCustomerEntityID(ctx context.Context, tx *sql.Tx) (string, error)
}

type idGeneratorService struct{}

func NewIDGeneratorService() IDGeneratorService {
	return &idGeneratorService{}
}

func (s *idGeneratorService) GenerateUUID() uuid.UUID {
	return uuid.New()
}

func (s *idGeneratorService) GenerateCustomerNumber(ctx context.Context, tx *sql.Tx, profileRepo repository.CustomerProfileRepository) (string, error) {
	// Use the context directly - the parent context already has a timeout
	// Nested timeouts can cause issues, so we rely on the transaction context timeout
	return profileRepo.GetNextCustomerNumber(ctx, tx)
}

func (s *idGeneratorService) GenerateCustomerEntityID(ctx context.Context, tx *sql.Tx) (string, error) {
	// Use the context directly - the parent context already has a timeout
	// Nested timeouts can cause issues, so we rely on the transaction context timeout
	var entityID string
	err := tx.QueryRowContext(ctx, `SELECT nextval('customer_entity_id_seq')::text`).Scan(&entityID)
	return entityID, err
}

