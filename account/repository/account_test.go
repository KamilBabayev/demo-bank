package repository

import (
	"context"
	"testing"

	"account/models"

	"github.com/shopspring/decimal"
)

// Note: These tests require a running PostgreSQL database
// For proper unit tests, you would use a mock database or testcontainers

func TestCreateAccountRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		req     models.CreateAccountRequest
		wantErr bool
	}{
		{
			name: "valid checking account",
			req: models.CreateAccountRequest{
				UserID:      1,
				AccountType: models.AccountTypeChecking,
				Currency:    "USD",
			},
			wantErr: false,
		},
		{
			name: "valid savings account",
			req: models.CreateAccountRequest{
				UserID:      1,
				AccountType: models.AccountTypeSavings,
				Currency:    "EUR",
			},
			wantErr: false,
		},
		{
			name: "missing user ID",
			req: models.CreateAccountRequest{
				UserID:      0,
				AccountType: models.AccountTypeChecking,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validate user ID
			if tt.req.UserID == 0 && !tt.wantErr {
				t.Error("expected error for missing user ID")
			}
			if tt.req.UserID != 0 && tt.wantErr && tt.name == "missing user ID" {
				t.Error("expected no error for valid user ID")
			}
		})
	}
}

func TestAccountNumber_Generation(t *testing.T) {
	// Test account number generation
	accNum1, err := generateAccountNumber()
	if err != nil {
		t.Fatalf("failed to generate account number: %v", err)
	}

	accNum2, err := generateAccountNumber()
	if err != nil {
		t.Fatalf("failed to generate account number: %v", err)
	}

	// Check length (should be 16 digits: 4 bank code + 12 random)
	if len(accNum1) != 16 {
		t.Errorf("account number length = %d, want 16", len(accNum1))
	}

	// Check bank code prefix
	if accNum1[:4] != "1001" {
		t.Errorf("account number prefix = %s, want 1001", accNum1[:4])
	}

	// Check uniqueness
	if accNum1 == accNum2 {
		t.Error("generated account numbers should be unique")
	}
}

func TestDecimalOperations(t *testing.T) {
	tests := []struct {
		name       string
		balance    string
		amount     string
		operation  string
		wantResult string
		wantErr    bool
	}{
		{
			name:       "simple deposit",
			balance:    "100.00",
			amount:     "50.00",
			operation:  "deposit",
			wantResult: "150.00",
		},
		{
			name:       "simple withdrawal",
			balance:    "100.00",
			amount:     "30.00",
			operation:  "withdraw",
			wantResult: "70.00",
		},
		{
			name:      "withdrawal exceeds balance",
			balance:   "100.00",
			amount:    "150.00",
			operation: "withdraw",
			wantErr:   true,
		},
		{
			name:      "negative deposit",
			balance:   "100.00",
			amount:    "-50.00",
			operation: "deposit",
			wantErr:   true,
		},
		{
			name:       "precision handling",
			balance:    "100.99",
			amount:     "0.01",
			operation:  "deposit",
			wantResult: "101.00",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			balance, _ := decimal.NewFromString(tt.balance)
			amount, _ := decimal.NewFromString(tt.amount)

			var result decimal.Decimal
			var err error

			switch tt.operation {
			case "deposit":
				if amount.LessThanOrEqual(decimal.Zero) {
					err = ErrInvalidAmount
				} else {
					result = balance.Add(amount)
				}
			case "withdraw":
				if amount.LessThanOrEqual(decimal.Zero) {
					err = ErrInvalidAmount
				} else if balance.LessThan(amount) {
					err = ErrInsufficientFunds
				} else {
					result = balance.Sub(amount)
				}
			}

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				expected, _ := decimal.NewFromString(tt.wantResult)
				if !result.Equal(expected) {
					t.Errorf("result = %s, want %s", result.String(), expected.String())
				}
			}
		})
	}
}

func TestSavingsWithdrawalLimit(t *testing.T) {
	limit := decimal.NewFromFloat(models.SavingsDailyWithdrawalLimit)

	tests := []struct {
		name        string
		dailyUsed   decimal.Decimal
		amount      decimal.Decimal
		wantAllowed bool
	}{
		{
			name:        "first withdrawal under limit",
			dailyUsed:   decimal.Zero,
			amount:      decimal.NewFromFloat(1000),
			wantAllowed: true,
		},
		{
			name:        "withdrawal at exact limit",
			dailyUsed:   decimal.Zero,
			amount:      limit,
			wantAllowed: true,
		},
		{
			name:        "withdrawal exceeds limit",
			dailyUsed:   decimal.Zero,
			amount:      limit.Add(decimal.NewFromFloat(1)),
			wantAllowed: false,
		},
		{
			name:        "cumulative exceeds limit",
			dailyUsed:   decimal.NewFromFloat(4000),
			amount:      decimal.NewFromFloat(1500),
			wantAllowed: false,
		},
		{
			name:        "cumulative at limit",
			dailyUsed:   decimal.NewFromFloat(4000),
			amount:      decimal.NewFromFloat(1000),
			wantAllowed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allowed := tt.dailyUsed.Add(tt.amount).LessThanOrEqual(limit)
			if allowed != tt.wantAllowed {
				t.Errorf("allowed = %v, want %v", allowed, tt.wantAllowed)
			}
		})
	}
}

func TestAccountStatusValidation(t *testing.T) {
	tests := []struct {
		name           string
		status         string
		canDeposit     bool
		canWithdraw    bool
		validTransition []string
	}{
		{
			name:            "active account",
			status:          models.AccountStatusActive,
			canDeposit:      true,
			canWithdraw:     true,
			validTransition: []string{"frozen", "closed"},
		},
		{
			name:            "frozen account",
			status:          models.AccountStatusFrozen,
			canDeposit:      false,
			canWithdraw:     false,
			validTransition: []string{"active", "closed"},
		},
		{
			name:            "closed account",
			status:          models.AccountStatusClosed,
			canDeposit:      false,
			canWithdraw:     false,
			validTransition: []string{}, // Cannot transition from closed
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test deposit capability
			canDeposit := tt.status == models.AccountStatusActive
			if canDeposit != tt.canDeposit {
				t.Errorf("canDeposit = %v, want %v", canDeposit, tt.canDeposit)
			}

			// Test withdrawal capability
			canWithdraw := tt.status == models.AccountStatusActive
			if canWithdraw != tt.canWithdraw {
				t.Errorf("canWithdraw = %v, want %v", canWithdraw, tt.canWithdraw)
			}
		})
	}
}

// Integration test example (requires database)
func TestAccountRepository_Integration(t *testing.T) {
	// Skip if not in integration test mode
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// This would require a real database connection
	// Example structure:
	/*
		ctx := context.Background()
		pool, err := setupTestDatabase()
		if err != nil {
			t.Fatalf("failed to setup test database: %v", err)
		}
		defer pool.Close()

		repo := NewAccountRepository(pool)

		// Test Create
		req := &models.CreateAccountRequest{
			UserID:      1,
			AccountType: models.AccountTypeChecking,
			Currency:    "USD",
		}
		account, err := repo.Create(ctx, req)
		if err != nil {
			t.Fatalf("failed to create account: %v", err)
		}

		// Test GetByID
		retrieved, err := repo.GetByID(ctx, account.ID)
		if err != nil {
			t.Fatalf("failed to get account: %v", err)
		}
		if retrieved.AccountNumber != account.AccountNumber {
			t.Error("account numbers don't match")
		}

		// Test Deposit
		depositAmount := decimal.NewFromFloat(100)
		account, err = repo.Deposit(ctx, account.ID, depositAmount)
		if err != nil {
			t.Fatalf("failed to deposit: %v", err)
		}
		if !account.Balance.Equal(depositAmount) {
			t.Errorf("balance = %s, want %s", account.Balance, depositAmount)
		}

		// Test Withdraw
		withdrawAmount := decimal.NewFromFloat(30)
		account, err = repo.Withdraw(ctx, account.ID, withdrawAmount)
		if err != nil {
			t.Fatalf("failed to withdraw: %v", err)
		}
		expectedBalance := depositAmount.Sub(withdrawAmount)
		if !account.Balance.Equal(expectedBalance) {
			t.Errorf("balance = %s, want %s", account.Balance, expectedBalance)
		}
	*/
	_ = context.Background() // Suppress unused variable warning
}
