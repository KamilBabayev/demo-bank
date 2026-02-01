package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	// Create Gin router
	router := gin.Default()

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "notification",
		})
	})

	// Notification endpoints
	api := router.Group("/api/notifications")
	{
		api.GET("", listNotifications)
		api.GET("/:id", getNotification)
		api.POST("", createNotification)
		api.PUT("/:id/read", markAsRead)
		api.DELETE("/:id", deleteNotification)
	}

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Notification service starting on port %s", port)
	router.Run(":" + port)
}

// Temporary handlers (will implement with DB later)
func listNotifications(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "List all notifications - to be implemented",
		"data":    []interface{}{},
	})
}

func getNotification(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{
		"message": "Get notification by ID - to be implemented",
		"id":      id,
	})
}

func createNotification(c *gin.Context) {
	c.JSON(http.StatusCreated, gin.H{
		"message": "Create notification - to be implemented",
	})
}

func markAsRead(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{
		"message": "Mark notification as read - to be implemented",
		"id":      id,
	})
}

func deleteNotification(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{
		"message": "Delete notification - to be implemented",
		"id":      id,
	})
}
