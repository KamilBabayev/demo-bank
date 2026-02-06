package repository

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"time"

	"card/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

var (
	ErrCardNotFound   = errors.New("card not found")
	ErrCardBlocked    = errors.New("card is blocked")
	ErrCardCancelled  = errors.New("card is cancelled")
	ErrCardExpired    = errors.New("card is expired")
	ErrInvalidInput   = errors.New("invalid input")
	ErrAccountNotOwned = errors.New("account not owned by user")
)

// Default limits
var (
	DefaultDailyLimit          = decimal.NewFromFloat(5000.00)
	DefaultMonthlyLimit        = decimal.NewFromFloat(50000.00)
	DefaultPerTransactionLimit = decimal.NewFromFloat(2000.00)
)

type CardRepository struct {
	db *pgxpool.Pool
}

func NewCardRepository(db *pgxpool.Pool) *CardRepository {
	return &CardRepository{db: db}
}

// generateCardNumber generates a 16-digit card number
func generateCardNumber() (string, error) {
	// Format: 4 digit issuer code + 12 random digits
	issuerCode := "4001" // Demo bank issuer code
	for i := 0; i < 12; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		issuerCode += n.String()
	}
	return issuerCode, nil
}

// hashCardNumber creates a SHA-256 hash of the card number for lookups
func hashCardNumber(cardNumber string) string {
	hash := sha256.Sum256([]byte(cardNumber))
	return hex.EncodeToString(hash[:])
}

// maskCardNumber masks all but the last 4 digits
func maskCardNumber(cardNumber string) string {
	if len(cardNumber) < 4 {
		return cardNumber
	}
	return "**** **** **** " + cardNumber[len(cardNumber)-4:]
}

// generateCVV generates a random 3-digit CVV
func generateCVV() (string, error) {
	cvv := ""
	for i := 0; i < 3; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		cvv += n.String()
	}
	return cvv, nil
}

// hashSecret creates a SHA-256 hash of a secret (CVV or PIN)
func hashSecret(secret string) string {
	hash := sha256.Sum256([]byte(secret))
	return hex.EncodeToString(hash[:])
}

// Create creates a new card
func (r *CardRepository) Create(ctx context.Context, req *models.CreateCardRequest) (*models.Card, error) {
	cardNumber, err := generateCardNumber()
	if err != nil {
		return nil, fmt.Errorf("failed to generate card number: %w", err)
	}

	cvv, err := generateCVV()
	if err != nil {
		return nil, fmt.Errorf("failed to generate CVV: %w", err)
	}

	// Card expires in 5 years
	now := time.Now()
	expirationMonth := int(now.Month())
	expirationYear := now.Year() + 5

	query := `
		INSERT INTO cards (
			account_id, card_number, card_number_hash, card_type, cardholder_name,
			expiration_month, expiration_year, cvv_hash,
			daily_limit, monthly_limit, per_transaction_limit
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, account_id, card_number, card_number_hash, card_type, cardholder_name,
		          expiration_month, expiration_year, status,
		          daily_limit, monthly_limit, per_transaction_limit,
		          daily_used, monthly_used, last_usage_date, created_at, updated_at
	`

	card := &models.Card{}
	var lastUsageDate *time.Time
	err = r.db.QueryRow(
		ctx, query,
		req.AccountID, maskCardNumber(cardNumber), hashCardNumber(cardNumber),
		req.CardType, req.CardholderName, expirationMonth, expirationYear,
		hashSecret(cvv), DefaultDailyLimit, DefaultMonthlyLimit, DefaultPerTransactionLimit,
	).Scan(
		&card.ID, &card.AccountID, &card.CardNumber, &card.CardNumberHash, &card.CardType,
		&card.CardholderName, &card.ExpirationMonth, &card.ExpirationYear, &card.Status,
		&card.DailyLimit, &card.MonthlyLimit, &card.PerTransactionLimit,
		&card.DailyUsed, &card.MonthlyUsed, &lastUsageDate, &card.CreatedAt, &card.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create card: %w", err)
	}

	card.LastUsageDate = lastUsageDate
	return card, nil
}

// GetByID retrieves a card by ID
func (r *CardRepository) GetByID(ctx context.Context, id int64) (*models.Card, error) {
	query := `
		SELECT id, account_id, card_number, card_number_hash, card_type, cardholder_name,
		       expiration_month, expiration_year, status,
		       daily_limit, monthly_limit, per_transaction_limit,
		       daily_used, monthly_used, last_usage_date, created_at, updated_at
		FROM cards
		WHERE id = $1
	`

	card := &models.Card{}
	var lastUsageDate *time.Time
	err := r.db.QueryRow(ctx, query, id).Scan(
		&card.ID, &card.AccountID, &card.CardNumber, &card.CardNumberHash, &card.CardType,
		&card.CardholderName, &card.ExpirationMonth, &card.ExpirationYear, &card.Status,
		&card.DailyLimit, &card.MonthlyLimit, &card.PerTransactionLimit,
		&card.DailyUsed, &card.MonthlyUsed, &lastUsageDate, &card.CreatedAt, &card.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrCardNotFound
		}
		return nil, fmt.Errorf("failed to get card: %w", err)
	}

	card.LastUsageDate = lastUsageDate
	return card, nil
}

// ListByAccountID retrieves all cards for an account
func (r *CardRepository) ListByAccountID(ctx context.Context, accountID int64, limit, offset int) (*models.CardListResponse, error) {
	var total int64
	countQuery := `SELECT COUNT(*) FROM cards WHERE account_id = $1`
	if err := r.db.QueryRow(ctx, countQuery, accountID).Scan(&total); err != nil {
		return nil, fmt.Errorf("failed to count cards: %w", err)
	}

	query := `
		SELECT id, account_id, card_number, card_number_hash, card_type, cardholder_name,
		       expiration_month, expiration_year, status,
		       daily_limit, monthly_limit, per_transaction_limit,
		       daily_used, monthly_used, last_usage_date, created_at, updated_at
		FROM cards
		WHERE account_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, accountID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list cards: %w", err)
	}
	defer rows.Close()

	cards := []models.Card{}
	for rows.Next() {
		var card models.Card
		var lastUsageDate *time.Time
		err := rows.Scan(
			&card.ID, &card.AccountID, &card.CardNumber, &card.CardNumberHash, &card.CardType,
			&card.CardholderName, &card.ExpirationMonth, &card.ExpirationYear, &card.Status,
			&card.DailyLimit, &card.MonthlyLimit, &card.PerTransactionLimit,
			&card.DailyUsed, &card.MonthlyUsed, &lastUsageDate, &card.CreatedAt, &card.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan card: %w", err)
		}
		card.LastUsageDate = lastUsageDate
		cards = append(cards, card)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating cards: %w", err)
	}

	return &models.CardListResponse{
		Cards: cards,
		Total: total,
	}, nil
}

// ListAll retrieves all cards (admin only)
func (r *CardRepository) ListAll(ctx context.Context, limit, offset int) (*models.CardListResponse, error) {
	var total int64
	countQuery := `SELECT COUNT(*) FROM cards`
	if err := r.db.QueryRow(ctx, countQuery).Scan(&total); err != nil {
		return nil, fmt.Errorf("failed to count cards: %w", err)
	}

	query := `
		SELECT id, account_id, card_number, card_number_hash, card_type, cardholder_name,
		       expiration_month, expiration_year, status,
		       daily_limit, monthly_limit, per_transaction_limit,
		       daily_used, monthly_used, last_usage_date, created_at, updated_at
		FROM cards
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list cards: %w", err)
	}
	defer rows.Close()

	cards := []models.Card{}
	for rows.Next() {
		var card models.Card
		var lastUsageDate *time.Time
		err := rows.Scan(
			&card.ID, &card.AccountID, &card.CardNumber, &card.CardNumberHash, &card.CardType,
			&card.CardholderName, &card.ExpirationMonth, &card.ExpirationYear, &card.Status,
			&card.DailyLimit, &card.MonthlyLimit, &card.PerTransactionLimit,
			&card.DailyUsed, &card.MonthlyUsed, &lastUsageDate, &card.CreatedAt, &card.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan card: %w", err)
		}
		card.LastUsageDate = lastUsageDate
		cards = append(cards, card)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating cards: %w", err)
	}

	return &models.CardListResponse{
		Cards: cards,
		Total: total,
	}, nil
}

// ListByUserAccounts retrieves all cards for accounts owned by a user
func (r *CardRepository) ListByUserAccounts(ctx context.Context, accountIDs []int64, limit, offset int) (*models.CardListResponse, error) {
	if len(accountIDs) == 0 {
		return &models.CardListResponse{Cards: []models.Card{}, Total: 0}, nil
	}

	var total int64
	countQuery := `SELECT COUNT(*) FROM cards WHERE account_id = ANY($1)`
	if err := r.db.QueryRow(ctx, countQuery, accountIDs).Scan(&total); err != nil {
		return nil, fmt.Errorf("failed to count cards: %w", err)
	}

	query := `
		SELECT id, account_id, card_number, card_number_hash, card_type, cardholder_name,
		       expiration_month, expiration_year, status,
		       daily_limit, monthly_limit, per_transaction_limit,
		       daily_used, monthly_used, last_usage_date, created_at, updated_at
		FROM cards
		WHERE account_id = ANY($1)
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, accountIDs, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list cards: %w", err)
	}
	defer rows.Close()

	cards := []models.Card{}
	for rows.Next() {
		var card models.Card
		var lastUsageDate *time.Time
		err := rows.Scan(
			&card.ID, &card.AccountID, &card.CardNumber, &card.CardNumberHash, &card.CardType,
			&card.CardholderName, &card.ExpirationMonth, &card.ExpirationYear, &card.Status,
			&card.DailyLimit, &card.MonthlyLimit, &card.PerTransactionLimit,
			&card.DailyUsed, &card.MonthlyUsed, &lastUsageDate, &card.CreatedAt, &card.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan card: %w", err)
		}
		card.LastUsageDate = lastUsageDate
		cards = append(cards, card)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating cards: %w", err)
	}

	return &models.CardListResponse{
		Cards: cards,
		Total: total,
	}, nil
}

// Update updates card limits or status
func (r *CardRepository) Update(ctx context.Context, id int64, req *models.UpdateCardRequest) (*models.Card, error) {
	// Start a transaction
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Get current card
	card, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Build dynamic update query
	updates := []string{}
	args := []interface{}{}
	argIdx := 1

	if req.Status != nil {
		updates = append(updates, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, *req.Status)
		argIdx++
	}
	if req.DailyLimit != nil {
		updates = append(updates, fmt.Sprintf("daily_limit = $%d", argIdx))
		args = append(args, *req.DailyLimit)
		argIdx++
	}
	if req.MonthlyLimit != nil {
		updates = append(updates, fmt.Sprintf("monthly_limit = $%d", argIdx))
		args = append(args, *req.MonthlyLimit)
		argIdx++
	}
	if req.PerTransactionLimit != nil {
		updates = append(updates, fmt.Sprintf("per_transaction_limit = $%d", argIdx))
		args = append(args, *req.PerTransactionLimit)
		argIdx++
	}

	if len(updates) == 0 {
		return card, nil
	}

	query := fmt.Sprintf(`
		UPDATE cards
		SET %s, updated_at = NOW()
		WHERE id = $%d
		RETURNING id, account_id, card_number, card_number_hash, card_type, cardholder_name,
		          expiration_month, expiration_year, status,
		          daily_limit, monthly_limit, per_transaction_limit,
		          daily_used, monthly_used, last_usage_date, created_at, updated_at
	`, joinStrings(updates, ", "), argIdx)
	args = append(args, id)

	updatedCard := &models.Card{}
	var lastUsageDate *time.Time
	err = tx.QueryRow(ctx, query, args...).Scan(
		&updatedCard.ID, &updatedCard.AccountID, &updatedCard.CardNumber, &updatedCard.CardNumberHash,
		&updatedCard.CardType, &updatedCard.CardholderName, &updatedCard.ExpirationMonth,
		&updatedCard.ExpirationYear, &updatedCard.Status, &updatedCard.DailyLimit,
		&updatedCard.MonthlyLimit, &updatedCard.PerTransactionLimit, &updatedCard.DailyUsed,
		&updatedCard.MonthlyUsed, &lastUsageDate, &updatedCard.CreatedAt, &updatedCard.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrCardNotFound
		}
		return nil, fmt.Errorf("failed to update card: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	updatedCard.LastUsageDate = lastUsageDate
	return updatedCard, nil
}

// Block blocks a card
func (r *CardRepository) Block(ctx context.Context, id int64) (*models.Card, error) {
	status := models.CardStatusBlocked
	return r.Update(ctx, id, &models.UpdateCardRequest{Status: &status})
}

// Unblock unblocks a card (sets status to active)
func (r *CardRepository) Unblock(ctx context.Context, id int64) (*models.Card, error) {
	status := models.CardStatusActive
	return r.Update(ctx, id, &models.UpdateCardRequest{Status: &status})
}

// Cancel cancels a card
func (r *CardRepository) Cancel(ctx context.Context, id int64) error {
	query := `UPDATE cards SET status = 'cancelled', updated_at = NOW() WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to cancel card: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrCardNotFound
	}

	return nil
}

// SetPIN sets or changes the card PIN
func (r *CardRepository) SetPIN(ctx context.Context, id int64, pin string) error {
	if len(pin) != 4 {
		return ErrInvalidInput
	}

	query := `UPDATE cards SET pin_hash = $1, updated_at = NOW() WHERE id = $2`

	result, err := r.db.Exec(ctx, query, hashSecret(pin), id)
	if err != nil {
		return fmt.Errorf("failed to set PIN: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrCardNotFound
	}

	return nil
}

// Helper function to join strings
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}
