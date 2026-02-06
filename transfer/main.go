package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"transfer/db"
	"transfer/kafka"
	"transfer/models"
	"transfer/repository"

	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

var (
	dbPool        *pgxpool.Pool
	transferRepo  *repository.TransferRepository
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
		Database: getEnv("DB_NAME", "transfer_db"),
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
	transferRepo = repository.NewTransferRepository(dbPool)

	// Initialize Kafka
	kafkaBrokers := strings.Split(getEnv("KAFKA_BROKERS", "localhost:9092"), ",")

	// Ensure topics exist
	kafka.EnsureTopicExists(kafkaBrokers, kafka.TopicTransferRequested)
	kafka.EnsureTopicExists(kafkaBrokers, kafka.TopicTransferCompleted)
	kafka.EnsureTopicExists(kafkaBrokers, kafka.TopicTransferFailed)

	// Initialize producer
	kafkaProducer = kafka.NewProducer(kafkaBrokers)
	defer kafkaProducer.Close()

	// Initialize consumer
	kafkaConsumer = kafka.NewConsumer(kafkaBrokers, "transfer-service", transferRepo)
	kafkaConsumer.Start(ctx)
	defer kafkaConsumer.Close()

	// Create Gin router
	router := gin.Default()

	// Health check endpoint
	router.GET("/health", healthCheck)

	// Transfer endpoints
	api := router.Group("/api/transfers")
	{
		api.GET("", listTransfers)
		api.GET("/:id", getTransfer)
		api.POST("", createTransfer)
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

	log.Printf("Transfer service starting on port %s", port)
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

	c.JSON(http.StatusOK, gin.H{
		"status":   "healthy",
		"service":  "transfer",
		"database": dbStatus,
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

func listTransfers(c *gin.Context) {
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

	var result *models.TransferListResponse

	if role == "admin" {
		// Admin can see all transfers
		result, err = transferRepo.ListAll(c.Request.Context(), limit, offset)
	} else {
		// Get user's accounts from account service
		accountIDs, accErr := getUserAccountIDs(userID, role)
		if accErr != nil {
			log.Printf("Failed to get user accounts: %v", accErr)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user accounts"})
			return
		}
		log.Printf("User %d has accounts: %v", userID, accountIDs)
		result, err = transferRepo.ListByAccountIDs(c.Request.Context(), accountIDs, limit, offset)
	}

	if err != nil {
		log.Printf("Failed to list transfers: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list transfers"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// getUserAccountIDs calls the account service to get the user's account IDs
func getUserAccountIDs(userID int64, role string) ([]int64, error) {
	accountServiceURL := getEnv("ACCOUNT_SERVICE_URL", "http://account.account.svc.cluster.local:8080")

	req, err := http.NewRequest("GET", accountServiceURL+"/api/accounts", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Pass user context headers
	req.Header.Set("X-User-ID", strconv.FormatInt(userID, 10))
	req.Header.Set("X-User-Role", role)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call account service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("account service returned status %d", resp.StatusCode)
	}

	var accountsResp struct {
		Accounts []struct {
			ID int64 `json:"id"`
		} `json:"accounts"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&accountsResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	accountIDs := make([]int64, len(accountsResp.Accounts))
	for i, acc := range accountsResp.Accounts {
		accountIDs[i] = acc.ID
	}

	return accountIDs, nil
}

func getTransfer(c *gin.Context) {
	_, role, err := getUserContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	transferID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid transfer ID"})
		return
	}

	transfer, err := transferRepo.GetByID(c.Request.Context(), transferID)
	if err != nil {
		if errors.Is(err, repository.ErrTransferNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "transfer not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get transfer"})
		return
	}

	// TODO: Verify user owns source or destination account for non-admin
	// For now, admin can view all, others need proper account verification
	if role != "admin" {
		// In production, verify user owns from_account_id or to_account_id
		// by calling the account service
	}

	c.JSON(http.StatusOK, transfer)
}

func createTransfer(c *gin.Context) {
	userID, role, err := getUserContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var req models.CreateTransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Amount.LessThanOrEqual(decimal.Zero) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "amount must be positive"})
		return
	}

	if req.FromAccountID == req.ToAccountID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "source and destination accounts cannot be the same"})
		return
	}

	// TODO: Verify user owns the source account by calling account service
	// For now, we trust the request (in production, this must be verified)
	_ = userID
	_ = role

	// Create transfer record
	transfer, err := transferRepo.Create(c.Request.Context(), &req)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrInvalidAmount):
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid amount"})
		case errors.Is(err, repository.ErrSameAccount):
			c.JSON(http.StatusBadRequest, gin.H{"error": "source and destination accounts cannot be the same"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create transfer"})
		}
		return
	}

	// Mark as processing
	transfer, err = transferRepo.MarkAsProcessing(c.Request.Context(), transfer.ID)
	if err != nil {
		log.Printf("Failed to mark transfer %d as processing: %v", transfer.ID, err)
	}

	// Publish event to Kafka
	if err := kafkaProducer.PublishTransferRequested(c.Request.Context(), transfer); err != nil {
		log.Printf("Failed to publish transfer event: %v", err)
		// Mark as failed since we couldn't process it
		transferRepo.MarkAsFailed(c.Request.Context(), transfer.ID, "failed to publish transfer event")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to initiate transfer"})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"message":      "transfer initiated",
		"transfer_id":  transfer.ID,
		"reference_id": transfer.ReferenceID,
		"status":       transfer.Status,
	})
}
