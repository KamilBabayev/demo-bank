package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"transfer/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

var (
	ErrTransferNotFound = errors.New("transfer not found")
	ErrInvalidAmount    = errors.New("invalid amount")
	ErrInvalidInput     = errors.New("invalid input")
	ErrSameAccount      = errors.New("source and destination accounts cannot be the same")
)

type TransferRepository struct {
	db *pgxpool.Pool
}

func NewTransferRepository(db *pgxpool.Pool) *TransferRepository {
	return &TransferRepository{db: db}
}

// Create creates a new transfer record
func (r *TransferRepository) Create(ctx context.Context, req *models.CreateTransferRequest) (*models.Transfer, error) {
	if req.Amount.LessThanOrEqual(decimal.Zero) {
		return nil, ErrInvalidAmount
	}

	if req.FromAccountID == req.ToAccountID {
		return nil, ErrSameAccount
	}

	currency := req.Currency
	if currency == "" {
		currency = "USD"
	}

	query := `
		INSERT INTO transfers (from_account_id, to_account_id, amount, currency, status)
		VALUES ($1, $2, $3, $4, 'pending')
		RETURNING id, reference_id, from_account_id, to_account_id, amount, currency, status,
		          failure_reason, created_at, updated_at, completed_at
	`

	transfer := &models.Transfer{}
	err := r.db.QueryRow(
		ctx, query,
		req.FromAccountID, req.ToAccountID, req.Amount, currency,
	).Scan(
		&transfer.ID, &transfer.ReferenceID, &transfer.FromAccountID, &transfer.ToAccountID,
		&transfer.Amount, &transfer.Currency, &transfer.Status,
		&transfer.FailureReason, &transfer.CreatedAt, &transfer.UpdatedAt, &transfer.CompletedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create transfer: %w", err)
	}

	return transfer, nil
}

// GetByID retrieves a transfer by ID
func (r *TransferRepository) GetByID(ctx context.Context, id int64) (*models.Transfer, error) {
	query := `
		SELECT id, reference_id, from_account_id, to_account_id, amount, currency, status,
		       failure_reason, created_at, updated_at, completed_at
		FROM transfers
		WHERE id = $1
	`

	transfer := &models.Transfer{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&transfer.ID, &transfer.ReferenceID, &transfer.FromAccountID, &transfer.ToAccountID,
		&transfer.Amount, &transfer.Currency, &transfer.Status,
		&transfer.FailureReason, &transfer.CreatedAt, &transfer.UpdatedAt, &transfer.CompletedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTransferNotFound
		}
		return nil, fmt.Errorf("failed to get transfer: %w", err)
	}

	return transfer, nil
}

// GetByReferenceID retrieves a transfer by reference ID
func (r *TransferRepository) GetByReferenceID(ctx context.Context, referenceID uuid.UUID) (*models.Transfer, error) {
	query := `
		SELECT id, reference_id, from_account_id, to_account_id, amount, currency, status,
		       failure_reason, created_at, updated_at, completed_at
		FROM transfers
		WHERE reference_id = $1
	`

	transfer := &models.Transfer{}
	err := r.db.QueryRow(ctx, query, referenceID).Scan(
		&transfer.ID, &transfer.ReferenceID, &transfer.FromAccountID, &transfer.ToAccountID,
		&transfer.Amount, &transfer.Currency, &transfer.Status,
		&transfer.FailureReason, &transfer.CreatedAt, &transfer.UpdatedAt, &transfer.CompletedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTransferNotFound
		}
		return nil, fmt.Errorf("failed to get transfer: %w", err)
	}

	return transfer, nil
}

// ListByAccountID retrieves all transfers for an account (as source or destination)
func (r *TransferRepository) ListByAccountID(ctx context.Context, accountID int64, limit, offset int) (*models.TransferListResponse, error) {
	var total int64
	countQuery := `SELECT COUNT(*) FROM transfers WHERE from_account_id = $1 OR to_account_id = $1`
	if err := r.db.QueryRow(ctx, countQuery, accountID).Scan(&total); err != nil {
		return nil, fmt.Errorf("failed to count transfers: %w", err)
	}

	query := `
		SELECT id, reference_id, from_account_id, to_account_id, amount, currency, status,
		       failure_reason, created_at, updated_at, completed_at
		FROM transfers
		WHERE from_account_id = $1 OR to_account_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, accountID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list transfers: %w", err)
	}
	defer rows.Close()

	transfers := []models.Transfer{}
	for rows.Next() {
		var transfer models.Transfer
		err := rows.Scan(
			&transfer.ID, &transfer.ReferenceID, &transfer.FromAccountID, &transfer.ToAccountID,
			&transfer.Amount, &transfer.Currency, &transfer.Status,
			&transfer.FailureReason, &transfer.CreatedAt, &transfer.UpdatedAt, &transfer.CompletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan transfer: %w", err)
		}
		transfers = append(transfers, transfer)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating transfers: %w", err)
	}

	return &models.TransferListResponse{
		Transfers: transfers,
		Total:     total,
	}, nil
}

// ListByAccountIDs retrieves all transfers for multiple accounts (as source or destination)
func (r *TransferRepository) ListByAccountIDs(ctx context.Context, accountIDs []int64, limit, offset int) (*models.TransferListResponse, error) {
	if len(accountIDs) == 0 {
		return &models.TransferListResponse{Transfers: []models.Transfer{}, Total: 0}, nil
	}

	// Build the IN clause - use same placeholders for both IN clauses (PostgreSQL allows reusing $1, $2, etc.)
	placeholders := ""
	args := make([]interface{}, len(accountIDs))
	for i, id := range accountIDs {
		if i > 0 {
			placeholders += ", "
		}
		placeholders += fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	var total int64
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM transfers WHERE from_account_id IN (%s) OR to_account_id IN (%s)`, placeholders, placeholders)
	// Only pass args once since we reuse the same placeholders
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("failed to count transfers: %w", err)
	}

	query := fmt.Sprintf(`
		SELECT id, reference_id, from_account_id, to_account_id, amount, currency, status,
		       failure_reason, created_at, updated_at, completed_at
		FROM transfers
		WHERE from_account_id IN (%s) OR to_account_id IN (%s)
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, placeholders, placeholders, len(accountIDs)+1, len(accountIDs)+2)

	queryArgs := append(args, limit, offset)
	rows, err := r.db.Query(ctx, query, queryArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to list transfers: %w", err)
	}
	defer rows.Close()

	transfers := []models.Transfer{}
	for rows.Next() {
		var transfer models.Transfer
		err := rows.Scan(
			&transfer.ID, &transfer.ReferenceID, &transfer.FromAccountID, &transfer.ToAccountID,
			&transfer.Amount, &transfer.Currency, &transfer.Status,
			&transfer.FailureReason, &transfer.CreatedAt, &transfer.UpdatedAt, &transfer.CompletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan transfer: %w", err)
		}
		transfers = append(transfers, transfer)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating transfers: %w", err)
	}

	return &models.TransferListResponse{
		Transfers: transfers,
		Total:     total,
	}, nil
}

// ListAll retrieves all transfers (admin only)
func (r *TransferRepository) ListAll(ctx context.Context, limit, offset int) (*models.TransferListResponse, error) {
	var total int64
	countQuery := `SELECT COUNT(*) FROM transfers`
	if err := r.db.QueryRow(ctx, countQuery).Scan(&total); err != nil {
		return nil, fmt.Errorf("failed to count transfers: %w", err)
	}

	query := `
		SELECT id, reference_id, from_account_id, to_account_id, amount, currency, status,
		       failure_reason, created_at, updated_at, completed_at
		FROM transfers
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list transfers: %w", err)
	}
	defer rows.Close()

	transfers := []models.Transfer{}
	for rows.Next() {
		var transfer models.Transfer
		err := rows.Scan(
			&transfer.ID, &transfer.ReferenceID, &transfer.FromAccountID, &transfer.ToAccountID,
			&transfer.Amount, &transfer.Currency, &transfer.Status,
			&transfer.FailureReason, &transfer.CreatedAt, &transfer.UpdatedAt, &transfer.CompletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan transfer: %w", err)
		}
		transfers = append(transfers, transfer)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating transfers: %w", err)
	}

	return &models.TransferListResponse{
		Transfers: transfers,
		Total:     total,
	}, nil
}

// UpdateStatus updates the transfer status
func (r *TransferRepository) UpdateStatus(ctx context.Context, id int64, status string, failureReason *string) (*models.Transfer, error) {
	var query string
	var args []interface{}

	if status == models.TransferStatusCompleted || status == models.TransferStatusFailed {
		query = `
			UPDATE transfers
			SET status = $1, failure_reason = $2, completed_at = $3, updated_at = NOW()
			WHERE id = $4
			RETURNING id, reference_id, from_account_id, to_account_id, amount, currency, status,
			          failure_reason, created_at, updated_at, completed_at
		`
		args = []interface{}{status, failureReason, time.Now(), id}
	} else {
		query = `
			UPDATE transfers
			SET status = $1, updated_at = NOW()
			WHERE id = $2
			RETURNING id, reference_id, from_account_id, to_account_id, amount, currency, status,
			          failure_reason, created_at, updated_at, completed_at
		`
		args = []interface{}{status, id}
	}

	transfer := &models.Transfer{}
	err := r.db.QueryRow(ctx, query, args...).Scan(
		&transfer.ID, &transfer.ReferenceID, &transfer.FromAccountID, &transfer.ToAccountID,
		&transfer.Amount, &transfer.Currency, &transfer.Status,
		&transfer.FailureReason, &transfer.CreatedAt, &transfer.UpdatedAt, &transfer.CompletedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTransferNotFound
		}
		return nil, fmt.Errorf("failed to update transfer status: %w", err)
	}

	return transfer, nil
}

// MarkAsProcessing marks a transfer as processing
func (r *TransferRepository) MarkAsProcessing(ctx context.Context, id int64) (*models.Transfer, error) {
	return r.UpdateStatus(ctx, id, models.TransferStatusProcessing, nil)
}

// MarkAsCompleted marks a transfer as completed
func (r *TransferRepository) MarkAsCompleted(ctx context.Context, id int64) (*models.Transfer, error) {
	return r.UpdateStatus(ctx, id, models.TransferStatusCompleted, nil)
}

// MarkAsFailed marks a transfer as failed
func (r *TransferRepository) MarkAsFailed(ctx context.Context, id int64, reason string) (*models.Transfer, error) {
	return r.UpdateStatus(ctx, id, models.TransferStatusFailed, &reason)
}
