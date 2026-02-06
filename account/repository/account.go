package repository

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"time"

	"account/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

var (
	ErrAccountNotFound       = errors.New("account not found")
	ErrInsufficientFunds     = errors.New("insufficient funds")
	ErrAccountFrozen         = errors.New("account is frozen")
	ErrAccountClosed         = errors.New("account is closed")
	ErrWithdrawalLimitExceed = errors.New("daily withdrawal limit exceeded for savings account")
	ErrInvalidAmount         = errors.New("invalid amount")
	ErrInvalidInput          = errors.New("invalid input")
)

type AccountRepository struct {
	db *pgxpool.Pool
}

func NewAccountRepository(db *pgxpool.Pool) *AccountRepository {
	return &AccountRepository{db: db}
}

// generateAccountNumber generates a unique 16-digit account number
func generateAccountNumber() (string, error) {
	// Format: 4 digit bank code + 12 random digits
	bankCode := "1001" // Demo bank code
	for i := 0; i < 12; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		bankCode += n.String()
	}
	return bankCode, nil
}

// Create creates a new account
func (r *AccountRepository) Create(ctx context.Context, req *models.CreateAccountRequest) (*models.Account, error) {
	accountNumber, err := generateAccountNumber()
	if err != nil {
		return nil, fmt.Errorf("failed to generate account number: %w", err)
	}

	currency := req.Currency
	if currency == "" {
		currency = "USD"
	}

	query := `
		INSERT INTO accounts (user_id, account_number, account_type, currency)
		VALUES ($1, $2, $3, $4)
		RETURNING id, user_id, account_number, account_type, balance, currency, status,
		          daily_withdrawal_used, last_withdrawal_date, created_at, updated_at
	`

	account := &models.Account{}
	err = r.db.QueryRow(
		ctx, query,
		req.UserID, accountNumber, req.AccountType, currency,
	).Scan(
		&account.ID, &account.UserID, &account.AccountNumber, &account.AccountType,
		&account.Balance, &account.Currency, &account.Status,
		&account.DailyWithdrawalUsed, &account.LastWithdrawalDate,
		&account.CreatedAt, &account.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create account: %w", err)
	}

	return account, nil
}

// GetByID retrieves an account by ID
func (r *AccountRepository) GetByID(ctx context.Context, id int64) (*models.Account, error) {
	query := `
		SELECT id, user_id, account_number, account_type, balance, currency, status,
		       daily_withdrawal_used, last_withdrawal_date, created_at, updated_at
		FROM accounts
		WHERE id = $1
	`

	account := &models.Account{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&account.ID, &account.UserID, &account.AccountNumber, &account.AccountType,
		&account.Balance, &account.Currency, &account.Status,
		&account.DailyWithdrawalUsed, &account.LastWithdrawalDate,
		&account.CreatedAt, &account.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrAccountNotFound
		}
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	return account, nil
}

// GetByAccountNumber retrieves an account by account number
func (r *AccountRepository) GetByAccountNumber(ctx context.Context, accountNumber string) (*models.Account, error) {
	query := `
		SELECT id, user_id, account_number, account_type, balance, currency, status,
		       daily_withdrawal_used, last_withdrawal_date, created_at, updated_at
		FROM accounts
		WHERE account_number = $1
	`

	account := &models.Account{}
	err := r.db.QueryRow(ctx, query, accountNumber).Scan(
		&account.ID, &account.UserID, &account.AccountNumber, &account.AccountType,
		&account.Balance, &account.Currency, &account.Status,
		&account.DailyWithdrawalUsed, &account.LastWithdrawalDate,
		&account.CreatedAt, &account.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrAccountNotFound
		}
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	return account, nil
}

// ListByUserID retrieves all accounts for a user
func (r *AccountRepository) ListByUserID(ctx context.Context, userID int64, limit, offset int) (*models.AccountListResponse, error) {
	var total int64
	countQuery := `SELECT COUNT(*) FROM accounts WHERE user_id = $1`
	if err := r.db.QueryRow(ctx, countQuery, userID).Scan(&total); err != nil {
		return nil, fmt.Errorf("failed to count accounts: %w", err)
	}

	query := `
		SELECT id, user_id, account_number, account_type, balance, currency, status,
		       daily_withdrawal_used, last_withdrawal_date, created_at, updated_at
		FROM accounts
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list accounts: %w", err)
	}
	defer rows.Close()

	accounts := []models.Account{}
	for rows.Next() {
		var account models.Account
		err := rows.Scan(
			&account.ID, &account.UserID, &account.AccountNumber, &account.AccountType,
			&account.Balance, &account.Currency, &account.Status,
			&account.DailyWithdrawalUsed, &account.LastWithdrawalDate,
			&account.CreatedAt, &account.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan account: %w", err)
		}
		accounts = append(accounts, account)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating accounts: %w", err)
	}

	return &models.AccountListResponse{
		Accounts: accounts,
		Total:    total,
	}, nil
}

// ListAllActive retrieves all active accounts (for transfer directory)
func (r *AccountRepository) ListAllActive(ctx context.Context) (*models.AccountListResponse, error) {
	query := `
		SELECT id, user_id, account_number, account_type, balance, currency, status,
		       daily_withdrawal_used, last_withdrawal_date, created_at, updated_at
		FROM accounts
		WHERE status = 'active'
		ORDER BY account_number ASC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list active accounts: %w", err)
	}
	defer rows.Close()

	accounts := []models.Account{}
	for rows.Next() {
		var account models.Account
		err := rows.Scan(
			&account.ID, &account.UserID, &account.AccountNumber, &account.AccountType,
			&account.Balance, &account.Currency, &account.Status,
			&account.DailyWithdrawalUsed, &account.LastWithdrawalDate,
			&account.CreatedAt, &account.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan account: %w", err)
		}
		accounts = append(accounts, account)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating accounts: %w", err)
	}

	return &models.AccountListResponse{
		Accounts: accounts,
		Total:    int64(len(accounts)),
	}, nil
}

// ListAll retrieves all accounts (admin only)
func (r *AccountRepository) ListAll(ctx context.Context, limit, offset int) (*models.AccountListResponse, error) {
	var total int64
	countQuery := `SELECT COUNT(*) FROM accounts`
	if err := r.db.QueryRow(ctx, countQuery).Scan(&total); err != nil {
		return nil, fmt.Errorf("failed to count accounts: %w", err)
	}

	query := `
		SELECT id, user_id, account_number, account_type, balance, currency, status,
		       daily_withdrawal_used, last_withdrawal_date, created_at, updated_at
		FROM accounts
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list accounts: %w", err)
	}
	defer rows.Close()

	accounts := []models.Account{}
	for rows.Next() {
		var account models.Account
		err := rows.Scan(
			&account.ID, &account.UserID, &account.AccountNumber, &account.AccountType,
			&account.Balance, &account.Currency, &account.Status,
			&account.DailyWithdrawalUsed, &account.LastWithdrawalDate,
			&account.CreatedAt, &account.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan account: %w", err)
		}
		accounts = append(accounts, account)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating accounts: %w", err)
	}

	return &models.AccountListResponse{
		Accounts: accounts,
		Total:    total,
	}, nil
}

// Update updates account status
func (r *AccountRepository) Update(ctx context.Context, id int64, req *models.UpdateAccountRequest) (*models.Account, error) {
	if req.Status == nil {
		return r.GetByID(ctx, id)
	}

	query := `
		UPDATE accounts
		SET status = $1, updated_at = NOW()
		WHERE id = $2
		RETURNING id, user_id, account_number, account_type, balance, currency, status,
		          daily_withdrawal_used, last_withdrawal_date, created_at, updated_at
	`

	account := &models.Account{}
	err := r.db.QueryRow(ctx, query, *req.Status, id).Scan(
		&account.ID, &account.UserID, &account.AccountNumber, &account.AccountType,
		&account.Balance, &account.Currency, &account.Status,
		&account.DailyWithdrawalUsed, &account.LastWithdrawalDate,
		&account.CreatedAt, &account.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrAccountNotFound
		}
		return nil, fmt.Errorf("failed to update account: %w", err)
	}

	return account, nil
}

// Delete soft deletes an account (sets status to closed)
func (r *AccountRepository) Delete(ctx context.Context, id int64) error {
	query := `UPDATE accounts SET status = 'closed', updated_at = NOW() WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete account: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrAccountNotFound
	}

	return nil
}

// Deposit adds funds to an account
func (r *AccountRepository) Deposit(ctx context.Context, id int64, amount decimal.Decimal) (*models.Account, error) {
	if amount.LessThanOrEqual(decimal.Zero) {
		return nil, ErrInvalidAmount
	}

	// Get account first to check status
	account, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if account.Status == models.AccountStatusFrozen {
		return nil, ErrAccountFrozen
	}
	if account.Status == models.AccountStatusClosed {
		return nil, ErrAccountClosed
	}

	query := `
		UPDATE accounts
		SET balance = balance + $1, updated_at = NOW()
		WHERE id = $2
		RETURNING id, user_id, account_number, account_type, balance, currency, status,
		          daily_withdrawal_used, last_withdrawal_date, created_at, updated_at
	`

	err = r.db.QueryRow(ctx, query, amount, id).Scan(
		&account.ID, &account.UserID, &account.AccountNumber, &account.AccountType,
		&account.Balance, &account.Currency, &account.Status,
		&account.DailyWithdrawalUsed, &account.LastWithdrawalDate,
		&account.CreatedAt, &account.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to deposit: %w", err)
	}

	return account, nil
}

// Withdraw removes funds from an account
func (r *AccountRepository) Withdraw(ctx context.Context, id int64, amount decimal.Decimal) (*models.Account, error) {
	if amount.LessThanOrEqual(decimal.Zero) {
		return nil, ErrInvalidAmount
	}

	// Use a transaction for atomic operations
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Lock the row for update
	query := `
		SELECT id, user_id, account_number, account_type, balance, currency, status,
		       daily_withdrawal_used, last_withdrawal_date, created_at, updated_at
		FROM accounts
		WHERE id = $1
		FOR UPDATE
	`

	account := &models.Account{}
	err = tx.QueryRow(ctx, query, id).Scan(
		&account.ID, &account.UserID, &account.AccountNumber, &account.AccountType,
		&account.Balance, &account.Currency, &account.Status,
		&account.DailyWithdrawalUsed, &account.LastWithdrawalDate,
		&account.CreatedAt, &account.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrAccountNotFound
		}
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	if account.Status == models.AccountStatusFrozen {
		return nil, ErrAccountFrozen
	}
	if account.Status == models.AccountStatusClosed {
		return nil, ErrAccountClosed
	}

	if account.Balance.LessThan(amount) {
		return nil, ErrInsufficientFunds
	}

	// Check savings account daily limit
	if account.AccountType == models.AccountTypeSavings {
		today := time.Now().Truncate(24 * time.Hour)
		dailyUsed := account.DailyWithdrawalUsed

		// Reset daily counter if last withdrawal was on a different day
		if account.LastWithdrawalDate == nil || account.LastWithdrawalDate.Truncate(24*time.Hour).Before(today) {
			dailyUsed = decimal.Zero
		}

		limitDecimal := decimal.NewFromFloat(models.SavingsDailyWithdrawalLimit)
		if dailyUsed.Add(amount).GreaterThan(limitDecimal) {
			return nil, ErrWithdrawalLimitExceed
		}

		// Update with new daily withdrawal tracking
		updateQuery := `
			UPDATE accounts
			SET balance = balance - $1,
			    daily_withdrawal_used = $2,
			    last_withdrawal_date = $3,
			    updated_at = NOW()
			WHERE id = $4
			RETURNING id, user_id, account_number, account_type, balance, currency, status,
			          daily_withdrawal_used, last_withdrawal_date, created_at, updated_at
		`

		err = tx.QueryRow(ctx, updateQuery, amount, dailyUsed.Add(amount), today, id).Scan(
			&account.ID, &account.UserID, &account.AccountNumber, &account.AccountType,
			&account.Balance, &account.Currency, &account.Status,
			&account.DailyWithdrawalUsed, &account.LastWithdrawalDate,
			&account.CreatedAt, &account.UpdatedAt,
		)
	} else {
		// Checking account - no daily limit
		updateQuery := `
			UPDATE accounts
			SET balance = balance - $1, updated_at = NOW()
			WHERE id = $2
			RETURNING id, user_id, account_number, account_type, balance, currency, status,
			          daily_withdrawal_used, last_withdrawal_date, created_at, updated_at
		`

		err = tx.QueryRow(ctx, updateQuery, amount, id).Scan(
			&account.ID, &account.UserID, &account.AccountNumber, &account.AccountType,
			&account.Balance, &account.Currency, &account.Status,
			&account.DailyWithdrawalUsed, &account.LastWithdrawalDate,
			&account.CreatedAt, &account.UpdatedAt,
		)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to withdraw: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return account, nil
}

// Transfer moves funds between accounts atomically
func (r *AccountRepository) Transfer(ctx context.Context, fromID, toID int64, amount decimal.Decimal) error {
	if amount.LessThanOrEqual(decimal.Zero) {
		return ErrInvalidAmount
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Lock both accounts in consistent order to prevent deadlocks
	var firstID, secondID int64
	if fromID < toID {
		firstID, secondID = fromID, toID
	} else {
		firstID, secondID = toID, fromID
	}

	// Lock first account
	var firstAccount models.Account
	lockQuery := `
		SELECT id, user_id, account_number, account_type, balance, currency, status,
		       daily_withdrawal_used, last_withdrawal_date, created_at, updated_at
		FROM accounts WHERE id = $1 FOR UPDATE
	`
	err = tx.QueryRow(ctx, lockQuery, firstID).Scan(
		&firstAccount.ID, &firstAccount.UserID, &firstAccount.AccountNumber, &firstAccount.AccountType,
		&firstAccount.Balance, &firstAccount.Currency, &firstAccount.Status,
		&firstAccount.DailyWithdrawalUsed, &firstAccount.LastWithdrawalDate,
		&firstAccount.CreatedAt, &firstAccount.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrAccountNotFound
		}
		return fmt.Errorf("failed to lock first account: %w", err)
	}

	// Lock second account
	var secondAccount models.Account
	err = tx.QueryRow(ctx, lockQuery, secondID).Scan(
		&secondAccount.ID, &secondAccount.UserID, &secondAccount.AccountNumber, &secondAccount.AccountType,
		&secondAccount.Balance, &secondAccount.Currency, &secondAccount.Status,
		&secondAccount.DailyWithdrawalUsed, &secondAccount.LastWithdrawalDate,
		&secondAccount.CreatedAt, &secondAccount.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrAccountNotFound
		}
		return fmt.Errorf("failed to lock second account: %w", err)
	}

	// Identify source and destination
	var fromAccount, toAccount *models.Account
	if firstID == fromID {
		fromAccount, toAccount = &firstAccount, &secondAccount
	} else {
		fromAccount, toAccount = &secondAccount, &firstAccount
	}

	// Validate source account
	if fromAccount.Status == models.AccountStatusFrozen {
		return ErrAccountFrozen
	}
	if fromAccount.Status == models.AccountStatusClosed {
		return ErrAccountClosed
	}
	if fromAccount.Balance.LessThan(amount) {
		return ErrInsufficientFunds
	}

	// Validate destination account
	if toAccount.Status == models.AccountStatusFrozen {
		return fmt.Errorf("destination account is frozen")
	}
	if toAccount.Status == models.AccountStatusClosed {
		return fmt.Errorf("destination account is closed")
	}

	// Check savings withdrawal limit for source
	if fromAccount.AccountType == models.AccountTypeSavings {
		today := time.Now().Truncate(24 * time.Hour)
		dailyUsed := fromAccount.DailyWithdrawalUsed

		if fromAccount.LastWithdrawalDate == nil || fromAccount.LastWithdrawalDate.Truncate(24*time.Hour).Before(today) {
			dailyUsed = decimal.Zero
		}

		limitDecimal := decimal.NewFromFloat(models.SavingsDailyWithdrawalLimit)
		if dailyUsed.Add(amount).GreaterThan(limitDecimal) {
			return ErrWithdrawalLimitExceed
		}

		// Update source with daily tracking
		_, err = tx.Exec(ctx, `
			UPDATE accounts
			SET balance = balance - $1, daily_withdrawal_used = $2, last_withdrawal_date = $3, updated_at = NOW()
			WHERE id = $4
		`, amount, dailyUsed.Add(amount), today, fromID)
	} else {
		// Debit source
		_, err = tx.Exec(ctx, `UPDATE accounts SET balance = balance - $1, updated_at = NOW() WHERE id = $2`, amount, fromID)
	}
	if err != nil {
		return fmt.Errorf("failed to debit source account: %w", err)
	}

	// Credit destination
	_, err = tx.Exec(ctx, `UPDATE accounts SET balance = balance + $1, updated_at = NOW() WHERE id = $2`, amount, toID)
	if err != nil {
		return fmt.Errorf("failed to credit destination account: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transfer: %w", err)
	}

	return nil
}
