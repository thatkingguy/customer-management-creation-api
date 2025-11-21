package service

import (
	"context"
	"database/sql"
	"time"

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
	// Add timeout to sequence call
	queryCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return profileRepo.GetNextCustomerNumber(queryCtx, tx)
}

func (s *idGeneratorService) GenerateCustomerEntityID(ctx context.Context, tx *sql.Tx) (string, error) {
	// Add timeout to sequence call
	queryCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	var entityID string
	err := tx.QueryRowContext(queryCtx, `SELECT nextval('customer_entity_id_seq')::text`).Scan(&entityID)
	return entityID, err
}

