package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"payment/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

var (
	ErrPaymentNotFound = errors.New("payment not found")
	ErrInvalidAmount   = errors.New("invalid amount")
	ErrInvalidInput    = errors.New("invalid input")
)

type PaymentRepository struct {
	db *pgxpool.Pool
}

func NewPaymentRepository(db *pgxpool.Pool) *PaymentRepository {
	return &PaymentRepository{db: db}
}

// Create creates a new payment record
func (r *PaymentRepository) Create(ctx context.Context, userID int64, req *models.CreatePaymentRequest) (*models.Payment, error) {
	if req.Amount.LessThanOrEqual(decimal.Zero) {
		return nil, ErrInvalidAmount
	}

	currency := req.Currency
	if currency == "" {
		currency = "USD"
	}

	query := `
		INSERT INTO payments (account_id, user_id, payment_type, recipient_name, recipient_account,
		                      recipient_bank, amount, currency, description, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, 'pending')
		RETURNING id, reference_id, account_id, user_id, payment_type, recipient_name, recipient_account,
		          recipient_bank, amount, currency, description, status, failure_reason,
		          created_at, updated_at, processed_at
	`

	payment := &models.Payment{}
	err := r.db.QueryRow(
		ctx, query,
		req.AccountID, userID, req.PaymentType, req.RecipientName, req.RecipientAccount,
		req.RecipientBank, req.Amount, currency, req.Description,
	).Scan(
		&payment.ID, &payment.ReferenceID, &payment.AccountID, &payment.UserID,
		&payment.PaymentType, &payment.RecipientName, &payment.RecipientAccount,
		&payment.RecipientBank, &payment.Amount, &payment.Currency, &payment.Description,
		&payment.Status, &payment.FailureReason, &payment.CreatedAt, &payment.UpdatedAt,
		&payment.ProcessedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create payment: %w", err)
	}

	return payment, nil
}

// GetByID retrieves a payment by ID
func (r *PaymentRepository) GetByID(ctx context.Context, id int64) (*models.Payment, error) {
	query := `
		SELECT id, reference_id, account_id, user_id, payment_type, recipient_name, recipient_account,
		       recipient_bank, amount, currency, description, status, failure_reason,
		       created_at, updated_at, processed_at
		FROM payments
		WHERE id = $1
	`

	payment := &models.Payment{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&payment.ID, &payment.ReferenceID, &payment.AccountID, &payment.UserID,
		&payment.PaymentType, &payment.RecipientName, &payment.RecipientAccount,
		&payment.RecipientBank, &payment.Amount, &payment.Currency, &payment.Description,
		&payment.Status, &payment.FailureReason, &payment.CreatedAt, &payment.UpdatedAt,
		&payment.ProcessedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrPaymentNotFound
		}
		return nil, fmt.Errorf("failed to get payment: %w", err)
	}

	return payment, nil
}

// GetByReferenceID retrieves a payment by reference ID
func (r *PaymentRepository) GetByReferenceID(ctx context.Context, referenceID uuid.UUID) (*models.Payment, error) {
	query := `
		SELECT id, reference_id, account_id, user_id, payment_type, recipient_name, recipient_account,
		       recipient_bank, amount, currency, description, status, failure_reason,
		       created_at, updated_at, processed_at
		FROM payments
		WHERE reference_id = $1
	`

	payment := &models.Payment{}
	err := r.db.QueryRow(ctx, query, referenceID).Scan(
		&payment.ID, &payment.ReferenceID, &payment.AccountID, &payment.UserID,
		&payment.PaymentType, &payment.RecipientName, &payment.RecipientAccount,
		&payment.RecipientBank, &payment.Amount, &payment.Currency, &payment.Description,
		&payment.Status, &payment.FailureReason, &payment.CreatedAt, &payment.UpdatedAt,
		&payment.ProcessedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrPaymentNotFound
		}
		return nil, fmt.Errorf("failed to get payment: %w", err)
	}

	return payment, nil
}

// ListByUserID retrieves all payments for a user
func (r *PaymentRepository) ListByUserID(ctx context.Context, userID int64, limit, offset int) (*models.PaymentListResponse, error) {
	var total int64
	countQuery := `SELECT COUNT(*) FROM payments WHERE user_id = $1`
	if err := r.db.QueryRow(ctx, countQuery, userID).Scan(&total); err != nil {
		return nil, fmt.Errorf("failed to count payments: %w", err)
	}

	query := `
		SELECT id, reference_id, account_id, user_id, payment_type, recipient_name, recipient_account,
		       recipient_bank, amount, currency, description, status, failure_reason,
		       created_at, updated_at, processed_at
		FROM payments
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list payments: %w", err)
	}
	defer rows.Close()

	payments := []models.Payment{}
	for rows.Next() {
		var payment models.Payment
		err := rows.Scan(
			&payment.ID, &payment.ReferenceID, &payment.AccountID, &payment.UserID,
			&payment.PaymentType, &payment.RecipientName, &payment.RecipientAccount,
			&payment.RecipientBank, &payment.Amount, &payment.Currency, &payment.Description,
			&payment.Status, &payment.FailureReason, &payment.CreatedAt, &payment.UpdatedAt,
			&payment.ProcessedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan payment: %w", err)
		}
		payments = append(payments, payment)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating payments: %w", err)
	}

	return &models.PaymentListResponse{
		Payments: payments,
		Total:    total,
	}, nil
}

// ListByAccountID retrieves all payments for an account
func (r *PaymentRepository) ListByAccountID(ctx context.Context, accountID int64, limit, offset int) (*models.PaymentListResponse, error) {
	var total int64
	countQuery := `SELECT COUNT(*) FROM payments WHERE account_id = $1`
	if err := r.db.QueryRow(ctx, countQuery, accountID).Scan(&total); err != nil {
		return nil, fmt.Errorf("failed to count payments: %w", err)
	}

	query := `
		SELECT id, reference_id, account_id, user_id, payment_type, recipient_name, recipient_account,
		       recipient_bank, amount, currency, description, status, failure_reason,
		       created_at, updated_at, processed_at
		FROM payments
		WHERE account_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, accountID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list payments: %w", err)
	}
	defer rows.Close()

	payments := []models.Payment{}
	for rows.Next() {
		var payment models.Payment
		err := rows.Scan(
			&payment.ID, &payment.ReferenceID, &payment.AccountID, &payment.UserID,
			&payment.PaymentType, &payment.RecipientName, &payment.RecipientAccount,
			&payment.RecipientBank, &payment.Amount, &payment.Currency, &payment.Description,
			&payment.Status, &payment.FailureReason, &payment.CreatedAt, &payment.UpdatedAt,
			&payment.ProcessedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan payment: %w", err)
		}
		payments = append(payments, payment)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating payments: %w", err)
	}

	return &models.PaymentListResponse{
		Payments: payments,
		Total:    total,
	}, nil
}

// ListAll retrieves all payments (admin only)
func (r *PaymentRepository) ListAll(ctx context.Context, limit, offset int) (*models.PaymentListResponse, error) {
	var total int64
	countQuery := `SELECT COUNT(*) FROM payments`
	if err := r.db.QueryRow(ctx, countQuery).Scan(&total); err != nil {
		return nil, fmt.Errorf("failed to count payments: %w", err)
	}

	query := `
		SELECT id, reference_id, account_id, user_id, payment_type, recipient_name, recipient_account,
		       recipient_bank, amount, currency, description, status, failure_reason,
		       created_at, updated_at, processed_at
		FROM payments
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list payments: %w", err)
	}
	defer rows.Close()

	payments := []models.Payment{}
	for rows.Next() {
		var payment models.Payment
		err := rows.Scan(
			&payment.ID, &payment.ReferenceID, &payment.AccountID, &payment.UserID,
			&payment.PaymentType, &payment.RecipientName, &payment.RecipientAccount,
			&payment.RecipientBank, &payment.Amount, &payment.Currency, &payment.Description,
			&payment.Status, &payment.FailureReason, &payment.CreatedAt, &payment.UpdatedAt,
			&payment.ProcessedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan payment: %w", err)
		}
		payments = append(payments, payment)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating payments: %w", err)
	}

	return &models.PaymentListResponse{
		Payments: payments,
		Total:    total,
	}, nil
}

// UpdateStatus updates the payment status
func (r *PaymentRepository) UpdateStatus(ctx context.Context, id int64, status string, failureReason *string) (*models.Payment, error) {
	var query string
	var args []interface{}

	if status == models.PaymentStatusCompleted || status == models.PaymentStatusFailed {
		query = `
			UPDATE payments
			SET status = $1, failure_reason = $2, processed_at = $3, updated_at = NOW()
			WHERE id = $4
			RETURNING id, reference_id, account_id, user_id, payment_type, recipient_name, recipient_account,
			          recipient_bank, amount, currency, description, status, failure_reason,
			          created_at, updated_at, processed_at
		`
		args = []interface{}{status, failureReason, time.Now(), id}
	} else {
		query = `
			UPDATE payments
			SET status = $1, updated_at = NOW()
			WHERE id = $2
			RETURNING id, reference_id, account_id, user_id, payment_type, recipient_name, recipient_account,
			          recipient_bank, amount, currency, description, status, failure_reason,
			          created_at, updated_at, processed_at
		`
		args = []interface{}{status, id}
	}

	payment := &models.Payment{}
	err := r.db.QueryRow(ctx, query, args...).Scan(
		&payment.ID, &payment.ReferenceID, &payment.AccountID, &payment.UserID,
		&payment.PaymentType, &payment.RecipientName, &payment.RecipientAccount,
		&payment.RecipientBank, &payment.Amount, &payment.Currency, &payment.Description,
		&payment.Status, &payment.FailureReason, &payment.CreatedAt, &payment.UpdatedAt,
		&payment.ProcessedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrPaymentNotFound
		}
		return nil, fmt.Errorf("failed to update payment status: %w", err)
	}

	return payment, nil
}

// MarkAsProcessing marks a payment as processing
func (r *PaymentRepository) MarkAsProcessing(ctx context.Context, id int64) (*models.Payment, error) {
	return r.UpdateStatus(ctx, id, models.PaymentStatusProcessing, nil)
}

// MarkAsCompleted marks a payment as completed
func (r *PaymentRepository) MarkAsCompleted(ctx context.Context, id int64) (*models.Payment, error) {
	return r.UpdateStatus(ctx, id, models.PaymentStatusCompleted, nil)
}

// MarkAsFailed marks a payment as failed
func (r *PaymentRepository) MarkAsFailed(ctx context.Context, id int64, reason string) (*models.Payment, error) {
	return r.UpdateStatus(ctx, id, models.PaymentStatusFailed, &reason)
}
