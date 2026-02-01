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
			"service": "payment",
		})
	})

	// Payment endpoints
	api := router.Group("/api/payments")
	{
		api.GET("", listPayments)
		api.GET("/:id", getPayment)
		api.POST("", createPayment)
		api.PUT("/:id/status", updatePaymentStatus)
		api.DELETE("/:id", deletePayment)
	}

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Payment service starting on port %s", port)
	router.Run(":" + port)
}

// Temporary handlers (will implement with DB later)
func listPayments(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "List all payments - to be implemented",
		"data":    []interface{}{},
	})
}

func getPayment(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{
		"message": "Get payment by ID - to be implemented",
		"id":      id,
	})
}

func createPayment(c *gin.Context) {
	c.JSON(http.StatusCreated, gin.H{
		"message": "Create payment - to be implemented",
	})
}

func updatePaymentStatus(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{
		"message": "Update payment status - to be implemented",
		"id":      id,
	})
}

func deletePayment(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{
		"message": "Delete payment - to be implemented",
		"id":      id,
	})
}
