package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"card/db"
	"card/kafka"
	"card/models"
	"card/repository"

	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	dbPool        *pgxpool.Pool
	cardRepo      *repository.CardRepository
	kafkaProducer *kafka.Producer
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
		Database: getEnv("DB_NAME", "card_db"),
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
	cardRepo = repository.NewCardRepository(dbPool)

	// Initialize Kafka
	kafkaBrokers := strings.Split(getEnv("KAFKA_BROKERS", "localhost:9092"), ",")

	// Ensure topics exist
	kafka.EnsureTopicExists(kafkaBrokers, kafka.TopicCardCreated)
	kafka.EnsureTopicExists(kafkaBrokers, kafka.TopicCardBlocked)
	kafka.EnsureTopicExists(kafkaBrokers, kafka.TopicCardActivated)
	kafka.EnsureTopicExists(kafkaBrokers, kafka.TopicCardCancelled)

	// Initialize producer
	kafkaProducer = kafka.NewProducer(kafkaBrokers)
	defer kafkaProducer.Close()

	// Create Gin router
	router := gin.Default()

	// Health check endpoint
	router.GET("/health", healthCheck)

	// Card endpoints
	api := router.Group("/api/cards")
	{
		api.GET("", listCards)
		api.GET("/:id", getCard)
		api.POST("", createCard)
		api.PUT("/:id", updateCard)
		api.DELETE("/:id", deleteCard)
		api.POST("/:id/block", blockCard)
		api.POST("/:id/unblock", unblockCard)
		api.POST("/:id/pin", setPIN)
		api.GET("/account/:accountId", listCardsByAccount)
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

	log.Printf("Card service starting on port %s", port)
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
		"service":  "card",
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

// getUserAccountIDs fetches all account IDs for a user from the account service
func getUserAccountIDs(ctx context.Context, userID int64) ([]int64, error) {
	accountServiceURL := getEnv("ACCOUNT_SERVICE_URL", "http://account.account.svc.cluster.local:8080")
	req, err := http.NewRequestWithContext(ctx, "GET", accountServiceURL+"/api/accounts", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-User-ID", strconv.FormatInt(userID, 10))
	req.Header.Set("X-User-Role", "customer")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("failed to fetch accounts")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result struct {
		Accounts []struct {
			ID int64 `json:"id"`
		} `json:"accounts"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	accountIDs := make([]int64, len(result.Accounts))
	for i, acc := range result.Accounts {
		accountIDs[i] = acc.ID
	}
	return accountIDs, nil
}

// checkAccountOwnership verifies that an account belongs to a user
func checkAccountOwnership(ctx context.Context, userID, accountID int64) (bool, error) {
	accountServiceURL := getEnv("ACCOUNT_SERVICE_URL", "http://account.account.svc.cluster.local:8080")
	req, err := http.NewRequestWithContext(ctx, "GET", accountServiceURL+"/api/accounts/"+strconv.FormatInt(accountID, 10), nil)
	if err != nil {
		return false, err
	}
	req.Header.Set("X-User-ID", strconv.FormatInt(userID, 10))
	req.Header.Set("X-User-Role", "customer")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	// If the account service returns 200, the user owns the account
	// If it returns 403 (forbidden) or 404 (not found), they don't
	return resp.StatusCode == http.StatusOK, nil
}

func listCards(c *gin.Context) {
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

	var result *models.CardListResponse

	if role == "admin" {
		// Admin can see all cards
		result, err = cardRepo.ListAll(c.Request.Context(), limit, offset)
	} else {
		// Customers can only see cards for their accounts
		accountIDs, accErr := getUserAccountIDs(c.Request.Context(), userID)
		if accErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch user accounts"})
			return
		}
		result, err = cardRepo.ListByUserAccounts(c.Request.Context(), accountIDs, limit, offset)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list cards"})
		return
	}

	c.JSON(http.StatusOK, result)
}

func getCard(c *gin.Context) {
	userID, role, err := getUserContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	cardID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid card ID"})
		return
	}

	card, err := cardRepo.GetByID(c.Request.Context(), cardID)
	if err != nil {
		if errors.Is(err, repository.ErrCardNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "card not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get card"})
		return
	}

	// Check ownership (non-admin can only view cards for their accounts)
	if role != "admin" {
		owns, err := checkAccountOwnership(c.Request.Context(), userID, card.AccountID)
		if err != nil || !owns {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}
	}

	c.JSON(http.StatusOK, card)
}

func createCard(c *gin.Context) {
	userID, role, err := getUserContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var req models.CreateCardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify account ownership for non-admin users
	if role != "admin" {
		owns, err := checkAccountOwnership(c.Request.Context(), userID, req.AccountID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to verify account ownership"})
			return
		}
		if !owns {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}
	}

	card, err := cardRepo.Create(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create card"})
		return
	}

	// Publish card created event
	event := models.CardEvent{
		CardID:         card.ID,
		AccountID:      card.AccountID,
		CardType:       card.CardType,
		CardholderName: card.CardholderName,
		Status:         card.Status,
	}
	if err := kafkaProducer.PublishCardCreated(c.Request.Context(), event); err != nil {
		log.Printf("Failed to publish card created event: %v", err)
	}

	c.JSON(http.StatusCreated, card)
}

func updateCard(c *gin.Context) {
	userID, role, err := getUserContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	cardID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid card ID"})
		return
	}

	// Get existing card to check ownership
	existingCard, err := cardRepo.GetByID(c.Request.Context(), cardID)
	if err != nil {
		if errors.Is(err, repository.ErrCardNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "card not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get card"})
		return
	}

	// Check ownership
	if role != "admin" {
		owns, err := checkAccountOwnership(c.Request.Context(), userID, existingCard.AccountID)
		if err != nil || !owns {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}
	}

	var req models.UpdateCardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Only admin can change status
	if req.Status != nil && role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "only admin can change card status"})
		return
	}

	card, err := cardRepo.Update(c.Request.Context(), cardID, &req)
	if err != nil {
		if errors.Is(err, repository.ErrCardNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "card not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update card"})
		return
	}

	c.JSON(http.StatusOK, card)
}

func deleteCard(c *gin.Context) {
	_, role, err := getUserContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Only admin can cancel cards
	if role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "only admin can cancel cards"})
		return
	}

	cardID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid card ID"})
		return
	}

	// Get card before cancelling for event
	card, err := cardRepo.GetByID(c.Request.Context(), cardID)
	if err != nil {
		if errors.Is(err, repository.ErrCardNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "card not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get card"})
		return
	}

	err = cardRepo.Cancel(c.Request.Context(), cardID)
	if err != nil {
		if errors.Is(err, repository.ErrCardNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "card not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to cancel card"})
		return
	}

	// Publish card cancelled event
	event := models.CardEvent{
		CardID:         card.ID,
		AccountID:      card.AccountID,
		CardType:       card.CardType,
		CardholderName: card.CardholderName,
		Status:         models.CardStatusCancelled,
	}
	if err := kafkaProducer.PublishCardCancelled(c.Request.Context(), event); err != nil {
		log.Printf("Failed to publish card cancelled event: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "card cancelled successfully"})
}

func blockCard(c *gin.Context) {
	userID, role, err := getUserContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	cardID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid card ID"})
		return
	}

	// Get existing card to check ownership
	existingCard, err := cardRepo.GetByID(c.Request.Context(), cardID)
	if err != nil {
		if errors.Is(err, repository.ErrCardNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "card not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get card"})
		return
	}

	// Check ownership (customers can block their own cards)
	if role != "admin" {
		owns, err := checkAccountOwnership(c.Request.Context(), userID, existingCard.AccountID)
		if err != nil || !owns {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}
	}

	card, err := cardRepo.Block(c.Request.Context(), cardID)
	if err != nil {
		if errors.Is(err, repository.ErrCardNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "card not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to block card"})
		return
	}

	// Publish card blocked event
	event := models.CardEvent{
		CardID:         card.ID,
		AccountID:      card.AccountID,
		CardType:       card.CardType,
		CardholderName: card.CardholderName,
		Status:         card.Status,
	}
	if err := kafkaProducer.PublishCardBlocked(c.Request.Context(), event); err != nil {
		log.Printf("Failed to publish card blocked event: %v", err)
	}

	c.JSON(http.StatusOK, card)
}

func unblockCard(c *gin.Context) {
	_, role, err := getUserContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Only admin can unblock cards
	if role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "only admin can unblock cards"})
		return
	}

	cardID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid card ID"})
		return
	}

	card, err := cardRepo.Unblock(c.Request.Context(), cardID)
	if err != nil {
		if errors.Is(err, repository.ErrCardNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "card not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to unblock card"})
		return
	}

	// Publish card activated event
	event := models.CardEvent{
		CardID:         card.ID,
		AccountID:      card.AccountID,
		CardType:       card.CardType,
		CardholderName: card.CardholderName,
		Status:         card.Status,
	}
	if err := kafkaProducer.PublishCardActivated(c.Request.Context(), event); err != nil {
		log.Printf("Failed to publish card activated event: %v", err)
	}

	c.JSON(http.StatusOK, card)
}

func setPIN(c *gin.Context) {
	userID, role, err := getUserContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	cardID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid card ID"})
		return
	}

	// Get existing card to check ownership
	existingCard, err := cardRepo.GetByID(c.Request.Context(), cardID)
	if err != nil {
		if errors.Is(err, repository.ErrCardNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "card not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get card"})
		return
	}

	// Check ownership
	if role != "admin" {
		owns, err := checkAccountOwnership(c.Request.Context(), userID, existingCard.AccountID)
		if err != nil || !owns {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}
	}

	var req models.SetPINRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = cardRepo.SetPIN(c.Request.Context(), cardID, req.PIN)
	if err != nil {
		if errors.Is(err, repository.ErrCardNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "card not found"})
			return
		}
		if errors.Is(err, repository.ErrInvalidInput) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "PIN must be 4 digits"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to set PIN"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "PIN set successfully"})
}

func listCardsByAccount(c *gin.Context) {
	userID, role, err := getUserContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	accountID, err := strconv.ParseInt(c.Param("accountId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid account ID"})
		return
	}

	// Check ownership for non-admin users
	if role != "admin" {
		owns, err := checkAccountOwnership(c.Request.Context(), userID, accountID)
		if err != nil || !owns {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	if limit > 100 {
		limit = 100
	}

	result, err := cardRepo.ListByAccountID(c.Request.Context(), accountID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list cards"})
		return
	}

	c.JSON(http.StatusOK, result)
}
