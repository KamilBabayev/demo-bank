package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"dbank/config"
)

type Client struct {
	httpClient *http.Client
	baseURL    string
	token      string
}

func NewClient(cfg *config.Config) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: cfg.APIURL,
		token:   cfg.Token,
	}
}

func (c *Client) doRequest(method, path string, body interface{}, result interface{}) error {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, c.baseURL+path, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		var errResp struct {
			Error string `json:"error"`
		}
		if json.Unmarshal(respBody, &errResp) == nil && errResp.Error != "" {
			return fmt.Errorf("API error: %s", errResp.Error)
		}
		return fmt.Errorf("API error: %s (status %d)", string(respBody), resp.StatusCode)
	}

	if result != nil {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}
	}

	return nil
}

// Auth endpoints

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  struct {
		ID       int64  `json:"id"`
		Username string `json:"username"`
		Role     string `json:"role"`
	} `json:"user"`
}

func (c *Client) Login(username, password string) (*LoginResponse, error) {
	var resp LoginResponse
	err := c.doRequest("POST", "/auth/login", LoginRequest{
		Username: username,
		Password: password,
	}, &resp)
	if err != nil {
		return nil, err
	}
	c.token = resp.Token
	return &resp, nil
}

// Account endpoints

type Account struct {
	ID            int64  `json:"id"`
	UserID        int64  `json:"user_id"`
	AccountNumber string `json:"account_number"`
	AccountType   string `json:"account_type"`
	Balance       string `json:"balance"`
	Currency      string `json:"currency"`
	Status        string `json:"status"`
	CreatedAt     string `json:"created_at"`
}

type AccountListResponse struct {
	Accounts []Account `json:"accounts"`
	Total    int64     `json:"total"`
}

func (c *Client) ListAccounts() (*AccountListResponse, error) {
	var resp AccountListResponse
	err := c.doRequest("GET", "/accounts", nil, &resp)
	return &resp, err
}

func (c *Client) GetAccount(id int64) (*Account, error) {
	var resp Account
	err := c.doRequest("GET", fmt.Sprintf("/accounts/%d", id), nil, &resp)
	return &resp, err
}

type DepositRequest struct {
	Amount   string `json:"amount"`
	Currency string `json:"currency,omitempty"`
}

func (c *Client) Deposit(accountID int64, amount string) (*Account, error) {
	var resp Account
	err := c.doRequest("POST", fmt.Sprintf("/accounts/%d/deposit", accountID), DepositRequest{
		Amount: amount,
	}, &resp)
	return &resp, err
}

func (c *Client) Withdraw(accountID int64, amount string) (*Account, error) {
	var resp Account
	err := c.doRequest("POST", fmt.Sprintf("/accounts/%d/withdraw", accountID), DepositRequest{
		Amount: amount,
	}, &resp)
	return &resp, err
}

// Transfer endpoints

type Transfer struct {
	ID            int64  `json:"id"`
	ReferenceID   string `json:"reference_id"`
	FromAccountID int64  `json:"from_account_id"`
	ToAccountID   int64  `json:"to_account_id"`
	Amount        string `json:"amount"`
	Currency      string `json:"currency"`
	Status        string `json:"status"`
	FailureReason string `json:"failure_reason,omitempty"`
	CreatedAt     string `json:"created_at"`
	CompletedAt   string `json:"completed_at,omitempty"`
}

type TransferListResponse struct {
	Transfers []Transfer `json:"transfers"`
	Total     int64      `json:"total"`
}

func (c *Client) ListTransfers() (*TransferListResponse, error) {
	var resp TransferListResponse
	err := c.doRequest("GET", "/transfers", nil, &resp)
	return &resp, err
}

type CreateTransferRequest struct {
	FromAccountID int64  `json:"from_account_id"`
	ToAccountID   int64  `json:"to_account_id"`
	Amount        string `json:"amount"`
	Currency      string `json:"currency,omitempty"`
}

type CreateTransferResponse struct {
	Message     string `json:"message"`
	TransferID  int64  `json:"transfer_id"`
	ReferenceID string `json:"reference_id"`
	Status      string `json:"status"`
}

func (c *Client) CreateTransfer(fromID, toID int64, amount, currency string) (*CreateTransferResponse, error) {
	var resp CreateTransferResponse
	req := CreateTransferRequest{
		FromAccountID: fromID,
		ToAccountID:   toID,
		Amount:        amount,
		Currency:      currency,
	}
	err := c.doRequest("POST", "/transfers", req, &resp)
	return &resp, err
}

func (c *Client) GetTransfer(id int64) (*Transfer, error) {
	var resp Transfer
	err := c.doRequest("GET", fmt.Sprintf("/transfers/%d", id), nil, &resp)
	return &resp, err
}

// Payment endpoints

type Payment struct {
	ID               int64  `json:"id"`
	ReferenceID      string `json:"reference_id"`
	AccountID        int64  `json:"account_id"`
	PaymentType      string `json:"payment_type"`
	RecipientName    string `json:"recipient_name,omitempty"`
	RecipientAccount string `json:"recipient_account,omitempty"`
	Amount           string `json:"amount"`
	Currency         string `json:"currency"`
	Status           string `json:"status"`
	FailureReason    string `json:"failure_reason,omitempty"`
	CreatedAt        string `json:"created_at"`
}

type PaymentListResponse struct {
	Payments []Payment `json:"payments"`
	Total    int64     `json:"total"`
}

func (c *Client) ListPayments() (*PaymentListResponse, error) {
	var resp PaymentListResponse
	err := c.doRequest("GET", "/payments", nil, &resp)
	return &resp, err
}

type CreatePaymentRequest struct {
	AccountID        int64  `json:"account_id"`
	PaymentType      string `json:"payment_type"`
	RecipientName    string `json:"recipient_name,omitempty"`
	RecipientAccount string `json:"recipient_account,omitempty"`
	Amount           string `json:"amount"`
	Currency         string `json:"currency,omitempty"`
	Description      string `json:"description,omitempty"`
}

func (c *Client) CreatePayment(req *CreatePaymentRequest) (*Payment, error) {
	var resp Payment
	err := c.doRequest("POST", "/payments", req, &resp)
	return &resp, err
}

// Notification endpoints

type Notification struct {
	ID        int64  `json:"id"`
	UserID    int64  `json:"user_id"`
	Type      string `json:"type"`
	Channel   string `json:"channel"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	Status    string `json:"status"`
	ReadAt    string `json:"read_at,omitempty"`
	CreatedAt string `json:"created_at"`
}

type NotificationListResponse struct {
	Notifications []Notification `json:"notifications"`
	Total         int64          `json:"total"`
	Unread        int64          `json:"unread"`
}

func (c *Client) ListNotifications() (*NotificationListResponse, error) {
	var resp NotificationListResponse
	err := c.doRequest("GET", "/notifications", nil, &resp)
	return &resp, err
}

func (c *Client) MarkNotificationRead(id int64) error {
	return c.doRequest("PUT", fmt.Sprintf("/notifications/%d/read", id), nil, nil)
}

func (c *Client) MarkAllNotificationsRead() error {
	return c.doRequest("PUT", "/notifications/read-all", nil, nil)
}

// User endpoints (admin)

type User struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Role      string `json:"role"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

type UserListResponse struct {
	Users []User `json:"users"`
	Total int64  `json:"total"`
}

func (c *Client) ListUsers() (*UserListResponse, error) {
	var resp UserListResponse
	err := c.doRequest("GET", "/users", nil, &resp)
	return &resp, err
}
