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

	"notification-service/cache"
	"notification-service/db"
	"notification-service/kafka"
	"notification-service/models"
	"notification-service/repository"

	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

var (
	dbPool           *pgxpool.Pool
	redisClient      *redis.Client
	notificationRepo repository.NotificationRepo
	kafkaConsumer    *kafka.Consumer
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
		Database: getEnv("DB_NAME", "notification_db"),
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
	baseRepo := repository.NewNotificationRepository(dbPool)

	// Initialize Redis cache
	redisURL := getEnv("REDIS_URL", "redis://localhost:6379")
	redisClient = cache.NewRedisClient(redisURL)
	defer redisClient.Close()

	notificationRepo = cache.NewCachedNotificationRepository(baseRepo, redisClient)

	// Initialize Kafka
	kafkaBrokers := strings.Split(getEnv("KAFKA_BROKERS", "localhost:9092"), ",")

	// Ensure topics exist (the notification service only consumes, doesn't produce)
	kafka.EnsureTopicExists(kafkaBrokers, kafka.TopicTransferCompleted)
	kafka.EnsureTopicExists(kafkaBrokers, kafka.TopicTransferFailed)
	kafka.EnsureTopicExists(kafkaBrokers, kafka.TopicPaymentCompleted)
	kafka.EnsureTopicExists(kafkaBrokers, kafka.TopicPaymentFailed)

	// Initialize consumer
	kafkaConsumer = kafka.NewConsumer(kafkaBrokers, "notification-service", notificationRepo)
	kafkaConsumer.Start(ctx)
	defer kafkaConsumer.Close()

	// Create Gin router
	router := gin.Default()

	// Health check endpoint
	router.GET("/health", healthCheck)

	// Notification endpoints
	api := router.Group("/api/notifications")
	{
		api.GET("", listNotifications)
		api.GET("/:id", getNotification)
		api.PUT("/:id/read", markAsRead)
		api.PUT("/read-all", markAllAsRead)
		api.DELETE("/:id", deleteNotification)
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

	log.Printf("Notification service starting on port %s", port)
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
		"service":  "notification",
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

func listNotifications(c *gin.Context) {
	userID, role, err := getUserContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	if limit > 100 {
		limit = 100
	}

	var result *models.NotificationListResponse

	if role == "admin" {
		// Admin can see all notifications
		result, err = notificationRepo.ListAll(c.Request.Context(), limit, offset)
	} else {
		// Users can see their own notifications
		result, err = notificationRepo.ListByUserID(c.Request.Context(), userID, limit, offset)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list notifications"})
		return
	}

	c.JSON(http.StatusOK, result)
}

func getNotification(c *gin.Context) {
	userID, role, err := getUserContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	notificationID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid notification ID"})
		return
	}

	notification, err := notificationRepo.GetByID(c.Request.Context(), notificationID)
	if err != nil {
		if errors.Is(err, repository.ErrNotificationNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "notification not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get notification"})
		return
	}

	// Check ownership (non-admin can only view own notifications)
	if role != "admin" && notification.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	c.JSON(http.StatusOK, notification)
}

func markAsRead(c *gin.Context) {
	userID, role, err := getUserContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	notificationID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid notification ID"})
		return
	}

	// Get notification to check ownership
	notification, err := notificationRepo.GetByID(c.Request.Context(), notificationID)
	if err != nil {
		if errors.Is(err, repository.ErrNotificationNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "notification not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get notification"})
		return
	}

	// Check ownership
	if role != "admin" && notification.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	notification, err = notificationRepo.MarkAsRead(c.Request.Context(), notificationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to mark notification as read"})
		return
	}

	c.JSON(http.StatusOK, notification)
}

func markAllAsRead(c *gin.Context) {
	userID, _, err := getUserContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	err = notificationRepo.MarkAllAsReadForUser(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to mark notifications as read"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "all notifications marked as read"})
}

func deleteNotification(c *gin.Context) {
	userID, role, err := getUserContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	notificationID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid notification ID"})
		return
	}

	// Get notification to check ownership
	notification, err := notificationRepo.GetByID(c.Request.Context(), notificationID)
	if err != nil {
		if errors.Is(err, repository.ErrNotificationNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "notification not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get notification"})
		return
	}

	// Check ownership
	if role != "admin" && notification.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	err = notificationRepo.Delete(c.Request.Context(), notificationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete notification"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "notification deleted"})
}
