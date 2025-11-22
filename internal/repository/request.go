package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/sterling-retailcore-team/customer-management-creation-api/internal/models"
)

type RequestRepository interface {
	Create(ctx context.Context, tx *sql.Tx, request *models.Request) error
	GetByID(ctx context.Context, db *sqlx.DB, requestID uuid.UUID) (*models.Request, error)
	Update(ctx context.Context, tx *sql.Tx, request *models.Request) error
	UpdateCustomerID(ctx context.Context, tx *sql.Tx, requestID uuid.UUID, customerID uuid.UUID) error
	UpdateStatus(ctx context.Context, tx *sql.Tx, requestID uuid.UUID, status models.RequestStatus, approvalStatus models.ApprovalStatus, approver *string, approverID *uuid.UUID, approverBranch *string) error
}

type requestRepository struct {
	db *sqlx.DB
}

func NewRequestRepository(db *sqlx.DB) RequestRepository {
	return &requestRepository{db: db}
}

func (r *requestRepository) Create(ctx context.Context, tx *sql.Tx, request *models.Request) error {
	query := `
		INSERT INTO request (
			"requestId", "customerId", "requestTitle", "requestType", "requestSubType",
			"accountNumber", "justification", "initiator", "initiatorId", "status",
			"approvalStatus", "approver", "approverId", "data", "customerType",
			"creationMode", "branch", "approverBranch", "department", "withdrawn",
			"isDeleted", "isProduct", "deletedOn", "rejectionDocument", "rejectionReason",
			"hasCollectionProduct", "bulkReferenceId", "createdAt", "updatedAt"
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29
		)
	`

	now := time.Now()
	request.CreatedAt = now
	request.UpdatedAt = now

	var dataJSON interface{}
	if request.Data != nil {
		dataJSON = request.Data
	}

	var rejectionDocJSON interface{}
	if request.RejectionDocument != nil {
		rejectionDocJSON = request.RejectionDocument
	}

	args := []interface{}{
		request.RequestID,
		request.CustomerID,
		request.RequestTitle,
		request.RequestType,
		request.RequestSubType,
		request.AccountNumber,
		request.Justification,
		request.Initiator,
		request.InitiatorID,
		request.Status,
		request.ApprovalStatus,
		request.Approver,
		request.ApproverID,
		dataJSON,
		request.CustomerType,
		request.CreationMode,
		request.Branch,
		request.ApproverBranch,
		request.Department,
		request.Withdrawn,
		request.IsDeleted,
		request.IsProduct,
		request.DeletedOn,
		rejectionDocJSON,
		request.RejectionReason,
		request.HasCollectionProduct,
		request.BulkReferenceID,
		request.CreatedAt,
		request.UpdatedAt,
	}

	// Use transaction if provided, otherwise use DB connection directly
	if tx != nil {
		_, err := tx.ExecContext(ctx, query, args...)
		return err
	}

	_, err := r.db.ExecContext(ctx, query, args...)
	return err
}

func (r *requestRepository) GetByID(ctx context.Context, db *sqlx.DB, requestID uuid.UUID) (*models.Request, error) {
	// Optimized query - using requestId (primary key) should be fast
	// The query itself executes in <1ms, but network latency adds overhead
	query := `
		SELECT 
			"requestId", "customerId", "requestTitle", "requestType", "requestSubType",
			"accountNumber", "justification", "initiator", "initiatorId", "status",
			"approvalStatus", "approver", "approverId", "data", "customerType",
			"creationMode", "branch", "approverBranch", "department", "withdrawn",
			"isDeleted", "isProduct", "deletedOn", "rejectionDocument", "rejectionReason",
			"hasCollectionProduct", "bulkReferenceId", "createdAt", "updatedAt"
		FROM request
		WHERE "requestId" = $1 AND "isDeleted" = false
	`

	// Use GetContext - sqlx will use prepared statements automatically for better performance
	// The 1-second delay is likely network latency, not query execution time
	var request models.Request
	err := db.GetContext(ctx, &request, query, requestID)
	if err != nil {
		return nil, err
	}

	return &request, nil
}

func (r *requestRepository) Update(ctx context.Context, tx *sql.Tx, request *models.Request) error {
	query := `
		UPDATE request SET
			"customerId" = $2, "requestTitle" = $3, "requestType" = $4, "requestSubType" = $5,
			"accountNumber" = $6, "justification" = $7, "initiator" = $8, "initiatorId" = $9,
			"status" = $10, "approvalStatus" = $11, "approver" = $12, "approverId" = $13,
			"data" = $14, "customerType" = $15, "creationMode" = $16, "branch" = $17,
			"approverBranch" = $18, "department" = $19, "withdrawn" = $20, "isDeleted" = $21,
			"isProduct" = $22, "deletedOn" = $23, "rejectionDocument" = $24, "rejectionReason" = $25,
			"hasCollectionProduct" = $26, "bulkReferenceId" = $27, "updatedAt" = $28
		WHERE "requestId" = $1
	`

	request.UpdatedAt = time.Now()

	var dataJSON interface{}
	if request.Data != nil {
		dataJSON = request.Data
	}

	var rejectionDocJSON interface{}
	if request.RejectionDocument != nil {
		rejectionDocJSON = request.RejectionDocument
	}

	_, err := tx.ExecContext(ctx, query,
		request.RequestID,
		request.CustomerID,
		request.RequestTitle,
		request.RequestType,
		request.RequestSubType,
		request.AccountNumber,
		request.Justification,
		request.Initiator,
		request.InitiatorID,
		request.Status,
		request.ApprovalStatus,
		request.Approver,
		request.ApproverID,
		dataJSON,
		request.CustomerType,
		request.CreationMode,
		request.Branch,
		request.ApproverBranch,
		request.Department,
		request.Withdrawn,
		request.IsDeleted,
		request.IsProduct,
		request.DeletedOn,
		rejectionDocJSON,
		request.RejectionReason,
		request.HasCollectionProduct,
		request.BulkReferenceID,
		request.UpdatedAt,
	)

	return err
}

func (r *requestRepository) UpdateCustomerID(ctx context.Context, tx *sql.Tx, requestID uuid.UUID, customerID uuid.UUID) error {
	query := `UPDATE request SET "customerId" = $2, "updatedAt" = $3 WHERE "requestId" = $1`
	args := []interface{}{requestID, customerID, time.Now()}
	
	// Use transaction if provided, otherwise use DB connection directly
	if tx != nil {
		_, err := tx.ExecContext(ctx, query, args...)
		return err
	}
	
	_, err := r.db.ExecContext(ctx, query, args...)
	return err
}

func (r *requestRepository) UpdateStatus(ctx context.Context, tx *sql.Tx, requestID uuid.UUID, status models.RequestStatus, approvalStatus models.ApprovalStatus, approver *string, approverID *uuid.UUID, approverBranch *string) error {
	query := `
		UPDATE request SET
			"status" = $2, "approvalStatus" = $3, "approver" = $4, "approverId" = $5, "approverBranch" = $6, "updatedAt" = $7
		WHERE "requestId" = $1
	`
	args := []interface{}{requestID, status, approvalStatus, approver, approverID, approverBranch, time.Now()}
	
	// Use transaction if provided, otherwise use DB connection directly
	if tx != nil {
		_, err := tx.ExecContext(ctx, query, args...)
		return err
	}
	
	_, err := r.db.ExecContext(ctx, query, args...)
	return err
}

