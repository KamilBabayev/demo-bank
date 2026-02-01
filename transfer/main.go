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
			"service": "transfer",
		})
	})

	// Transfer endpoints
	api := router.Group("/api/transfers")
	{
		api.GET("", listTransfers)
		api.GET("/:id", getTransfer)
		api.POST("", createTransfer)
		api.PUT("/:id/status", updateTransferStatus)
		api.DELETE("/:id", deleteTransfer)
	}

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Transfer service starting on port %s", port)
	router.Run(":" + port)
}

// Temporary handlers (will implement with DB later)
func listTransfers(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "List all transfers - to be implemented",
		"data":    []interface{}{},
	})
}

func getTransfer(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{
		"message": "Get transfer by ID - to be implemented",
		"id":      id,
	})
}

func createTransfer(c *gin.Context) {
	c.JSON(http.StatusCreated, gin.H{
		"message": "Create transfer - to be implemented",
	})
}

func updateTransferStatus(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{
		"message": "Update transfer status - to be implemented",
		"id":      id,
	})
}

func deleteTransfer(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{
		"message": "Delete transfer - to be implemented",
		"id":      id,
	})
}
