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

	"payment/cache"
	"payment/db"
	"payment/kafka"
	"payment/models"
	"payment/repository"

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
	paymentRepo   repository.PaymentRepo
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
		Database: getEnv("DB_NAME", "payment_db"),
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
	baseRepo := repository.NewPaymentRepository(dbPool)

	// Initialize Redis cache
	redisURL := getEnv("REDIS_URL", "redis://localhost:6379")
	redisClient = cache.NewRedisClient(redisURL)
	defer redisClient.Close()

	paymentRepo = cache.NewCachedPaymentRepository(baseRepo, redisClient)

	// Initialize Kafka
	kafkaBrokers := strings.Split(getEnv("KAFKA_BROKERS", "localhost:9092"), ",")

	// Ensure topics exist
	kafka.EnsureTopicExists(kafkaBrokers, kafka.TopicPaymentRequested)
	kafka.EnsureTopicExists(kafkaBrokers, kafka.TopicPaymentCompleted)
	kafka.EnsureTopicExists(kafkaBrokers, kafka.TopicPaymentFailed)

	// Initialize producer
	kafkaProducer = kafka.NewProducer(kafkaBrokers)
	defer kafkaProducer.Close()

	// Initialize consumer
	kafkaConsumer = kafka.NewConsumer(kafkaBrokers, "payment-service", paymentRepo)
	kafkaConsumer.Start(ctx)
	defer kafkaConsumer.Close()

	// Create Gin router
	router := gin.Default()

	// Health check endpoint
	router.GET("/health", healthCheck)

	// Payment endpoints
	api := router.Group("/api/payments")
	{
		api.GET("", listPayments)
		api.GET("/:id", getPayment)
		api.POST("", createPayment)
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

	log.Printf("Payment service starting on port %s", port)
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
		"service":  "payment",
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

func listPayments(c *gin.Context) {
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

	var result *models.PaymentListResponse

	if role == "admin" {
		// Admin can see all payments
		result, err = paymentRepo.ListAll(c.Request.Context(), limit, offset)
	} else {
		// Users can see their own payments
		result, err = paymentRepo.ListByUserID(c.Request.Context(), userID, limit, offset)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list payments"})
		return
	}

	c.JSON(http.StatusOK, result)
}

func getPayment(c *gin.Context) {
	userID, role, err := getUserContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	paymentID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payment ID"})
		return
	}

	payment, err := paymentRepo.GetByID(c.Request.Context(), paymentID)
	if err != nil {
		if errors.Is(err, repository.ErrPaymentNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "payment not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get payment"})
		return
	}

	// Check ownership (non-admin can only view own payments)
	if role != "admin" && payment.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	c.JSON(http.StatusOK, payment)
}

func createPayment(c *gin.Context) {
	userID, _, err := getUserContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var req models.CreatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Amount.LessThanOrEqual(decimal.Zero) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "amount must be positive"})
		return
	}

	// Validate based on payment type
	switch req.PaymentType {
	case models.PaymentTypeBill:
		if req.RecipientName == nil || *req.RecipientName == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "recipient_name required for bill payments"})
			return
		}
	case models.PaymentTypeMerchant:
		if req.RecipientName == nil || *req.RecipientName == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "recipient_name required for merchant payments"})
			return
		}
	case models.PaymentTypeExternal:
		if req.RecipientAccount == nil || *req.RecipientAccount == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "recipient_account required for external transfers"})
			return
		}
	}

	// TODO: Verify user owns the source account by calling account service

	// Create payment record
	payment, err := paymentRepo.Create(c.Request.Context(), userID, &req)
	if err != nil {
		if errors.Is(err, repository.ErrInvalidAmount) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid amount"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create payment"})
		return
	}

	// Mark as processing
	payment, err = paymentRepo.MarkAsProcessing(c.Request.Context(), payment.ID)
	if err != nil {
		log.Printf("Failed to mark payment %d as processing: %v", payment.ID, err)
	}

	// Publish event to Kafka
	if err := kafkaProducer.PublishPaymentRequested(c.Request.Context(), payment); err != nil {
		log.Printf("Failed to publish payment event: %v", err)
		// Mark as failed since we couldn't process it
		paymentRepo.MarkAsFailed(c.Request.Context(), payment.ID, "failed to publish payment event")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to initiate payment"})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"message":      "payment initiated",
		"payment_id":   payment.ID,
		"reference_id": payment.ReferenceID,
		"status":       payment.Status,
	})
}
