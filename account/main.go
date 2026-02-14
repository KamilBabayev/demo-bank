package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"account/cache"
	"account/db"
	"account/kafka"
	"account/models"
	"account/repository"

	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
)

var (
	dbPool        *pgxpool.Pool
	redisClient   *redis.Client
	accountRepo   repository.AccountRepo
	kafkaProducer *kafka.Producer
	kafkaConsumer *kafka.Consumer
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Database configuration
	dbConfig := db.Config{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnv("DB_PORT", "5432"),
		User:     getEnv("DB_USER", "postgres"),
		Password: getEnv("DB_PASSWORD", "postgres"),
		Database: getEnv("DB_NAME", "account_db"),
	}

	// Connect to database
	var err error
	dbPool, err = db.NewPool(ctx, dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbPool.Close()

	// Run migrations
	dbURL := "postgres://" + dbConfig.User + ":" + dbConfig.Password + "@" + dbConfig.Host + ":" + dbConfig.Port + "/" + dbConfig.Database + "?sslmode=disable"
	m, err := migrate.New("file://migrations", dbURL)
	if err != nil {
		log.Fatalf("Failed to create migrate instance: %v", err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Failed to run migrations: %v", err)
	}
	log.Println("Database migrations completed")

	// Initialize repository
	baseRepo := repository.NewAccountRepository(dbPool)

	// Initialize Redis cache
	redisURL := getEnv("REDIS_URL", "redis://localhost:6379")
	redisClient = cache.NewRedisClient(redisURL)
	defer redisClient.Close()

	accountRepo = cache.NewCachedAccountRepository(baseRepo, redisClient)

	// Initialize Kafka
	kafkaBrokers := strings.Split(getEnv("KAFKA_BROKERS", "localhost:9092"), ",")

	// Ensure topics exist
	kafka.EnsureTopicExists(kafkaBrokers, kafka.TopicTransferRequested)
	kafka.EnsureTopicExists(kafkaBrokers, kafka.TopicTransferCompleted)
	kafka.EnsureTopicExists(kafkaBrokers, kafka.TopicTransferFailed)
	kafka.EnsureTopicExists(kafkaBrokers, kafka.TopicPaymentRequested)
	kafka.EnsureTopicExists(kafkaBrokers, kafka.TopicPaymentCompleted)
	kafka.EnsureTopicExists(kafkaBrokers, kafka.TopicPaymentFailed)

	// Initialize producer
	kafkaProducer = kafka.NewProducer(kafkaBrokers)
	defer kafkaProducer.Close()

	// Initialize consumer
	kafkaConsumer = kafka.NewConsumer(kafkaBrokers, "account-service", accountRepo, kafkaProducer)
	go kafkaConsumer.Start(ctx)
	defer kafkaConsumer.Close()

	// Create Gin router
	router := gin.Default()

	// Health check endpoint
	router.GET("/health", healthCheck)

	// Account endpoints
	api := router.Group("/api/accounts")
	{
		api.GET("", listAccounts)
		api.GET("/directory", listAccountDirectory)
		api.GET("/:id", getAccount)
		api.POST("", createAccount)
		api.PUT("/:id", updateAccount)
		api.DELETE("/:id", deleteAccount)
		api.GET("/:id/balance", getBalance)
		api.POST("/:id/deposit", deposit)
		api.POST("/:id/withdraw", withdraw)
	}

	// Get port from environment or use default
	port := getEnv("PORT", "8080")

	// Graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Println("Shutting down...")
		cancel()
	}()

	log.Printf("Account service starting on port %s", port)
	router.Run(":" + port)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func healthCheck(c *gin.Context) {
	dbStatus := "connected"
	if err := db.HealthCheck(c.Request.Context(), dbPool); err != nil {
		dbStatus = "disconnected"
	}

	redisStatus := "connected"
	if err := cache.HealthCheck(c.Request.Context(), redisClient); err != nil {
		redisStatus = "disconnected"
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   "healthy",
		"service":  "account",
		"database": dbStatus,
		"redis":    redisStatus,
	})
}

// getUserContext extracts user ID and role from headers (set by API Gateway)
func getUserContext(c *gin.Context) (userID int64, role string, err error) {
	userIDStr := c.GetHeader("X-User-ID")
	role = c.GetHeader("X-User-Role")

	if userIDStr == "" {
		return 0, "", errors.New("missing user context")
	}

	userID, err = strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		return 0, "", errors.New("invalid user ID")
	}

	return userID, role, nil
}

func listAccounts(c *gin.Context) {
	userID, role, err := getUserContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	if limit > 100 {
		limit = 100
	}

	var result *models.AccountListResponse

	if role == "admin" {
		// Admin can see all accounts
		result, err = accountRepo.ListAll(c.Request.Context(), limit, offset)
	} else {
		// Customers can only see their own accounts
		result, err = accountRepo.ListByUserID(c.Request.Context(), userID, limit, offset)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list accounts"})
		return
	}

	c.JSON(http.StatusOK, result)
}

func listAccountDirectory(c *gin.Context) {
	_, _, err := getUserContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	result, err := accountRepo.ListAllActive(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list accounts"})
		return
	}

	c.JSON(http.StatusOK, result)
}

func getAccount(c *gin.Context) {
	userID, role, err := getUserContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	accountID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid account ID"})
		return
	}

	account, err := accountRepo.GetByID(c.Request.Context(), accountID)
	if err != nil {
		if errors.Is(err, repository.ErrAccountNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "account not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get account"})
		return
	}

	// Check ownership (non-admin can only view own accounts)
	if role != "admin" && account.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	c.JSON(http.StatusOK, account)
}

func createAccount(c *gin.Context) {
	userID, role, err := getUserContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var req models.CreateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Admin can create accounts for any user, customers can only create for themselves
	if role != "admin" {
		req.UserID = userID
	}

	account, err := accountRepo.Create(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create account"})
		return
	}

	c.JSON(http.StatusCreated, account)
}

func updateAccount(c *gin.Context) {
	userID, role, err := getUserContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	accountID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid account ID"})
		return
	}

	// Get existing account to check ownership
	existingAccount, err := accountRepo.GetByID(c.Request.Context(), accountID)
	if err != nil {
		if errors.Is(err, repository.ErrAccountNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "account not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get account"})
		return
	}

	// Only admin can update account status
	if role != "admin" && existingAccount.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	var req models.UpdateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Only admin can change status
	if req.Status != nil && role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "only admin can change account status"})
		return
	}

	account, err := accountRepo.Update(c.Request.Context(), accountID, &req)
	if err != nil {
		if errors.Is(err, repository.ErrAccountNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "account not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update account"})
		return
	}

	c.JSON(http.StatusOK, account)
}

func deleteAccount(c *gin.Context) {
	_, role, err := getUserContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Only admin can close accounts
	if role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "only admin can close accounts"})
		return
	}

	accountID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid account ID"})
		return
	}

	err = accountRepo.Delete(c.Request.Context(), accountID)
	if err != nil {
		if errors.Is(err, repository.ErrAccountNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "account not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to close account"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "account closed successfully"})
}

func getBalance(c *gin.Context) {
	userID, role, err := getUserContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	accountID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid account ID"})
		return
	}

	account, err := accountRepo.GetByID(c.Request.Context(), accountID)
	if err != nil {
		if errors.Is(err, repository.ErrAccountNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "account not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get account"})
		return
	}

	// Check ownership
	if role != "admin" && account.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	c.JSON(http.StatusOK, models.BalanceResponse{
		AccountID:     account.ID,
		AccountNumber: account.AccountNumber,
		Balance:       account.Balance,
		Currency:      account.Currency,
		AccountType:   account.AccountType,
		Status:        account.Status,
	})
}

func deposit(c *gin.Context) {
	userID, role, err := getUserContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	accountID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid account ID"})
		return
	}

	// Verify account exists and check ownership
	existingAccount, err := accountRepo.GetByID(c.Request.Context(), accountID)
	if err != nil {
		if errors.Is(err, repository.ErrAccountNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "account not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get account"})
		return
	}

	if role != "admin" && existingAccount.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	var req models.DepositRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Amount.LessThanOrEqual(decimal.Zero) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "amount must be positive"})
		return
	}

	account, err := accountRepo.Deposit(c.Request.Context(), accountID, req.Amount)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrAccountNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "account not found"})
		case errors.Is(err, repository.ErrAccountFrozen):
			c.JSON(http.StatusForbidden, gin.H{"error": "account is frozen"})
		case errors.Is(err, repository.ErrAccountClosed):
			c.JSON(http.StatusForbidden, gin.H{"error": "account is closed"})
		case errors.Is(err, repository.ErrInvalidAmount):
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid amount"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to deposit"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "deposit successful",
		"account": account,
	})
}

func withdraw(c *gin.Context) {
	userID, role, err := getUserContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	accountID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid account ID"})
		return
	}

	// Verify account exists and check ownership
	existingAccount, err := accountRepo.GetByID(c.Request.Context(), accountID)
	if err != nil {
		if errors.Is(err, repository.ErrAccountNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "account not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get account"})
		return
	}

	if role != "admin" && existingAccount.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	var req models.WithdrawRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Amount.LessThanOrEqual(decimal.Zero) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "amount must be positive"})
		return
	}

	account, err := accountRepo.Withdraw(c.Request.Context(), accountID, req.Amount)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrAccountNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "account not found"})
		case errors.Is(err, repository.ErrAccountFrozen):
			c.JSON(http.StatusForbidden, gin.H{"error": "account is frozen"})
		case errors.Is(err, repository.ErrAccountClosed):
			c.JSON(http.StatusForbidden, gin.H{"error": "account is closed"})
		case errors.Is(err, repository.ErrInsufficientFunds):
			c.JSON(http.StatusBadRequest, gin.H{"error": "insufficient funds"})
		case errors.Is(err, repository.ErrWithdrawalLimitExceed):
			c.JSON(http.StatusBadRequest, gin.H{"error": "daily withdrawal limit exceeded for savings account"})
		case errors.Is(err, repository.ErrInvalidAmount):
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid amount"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to withdraw"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "withdrawal successful",
		"account": account,
	})
}
