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
			"service": "account",
		})
	})

	// Account endpoints
	api := router.Group("/api/accounts")
	{
		api.GET("", listAccounts)
		api.GET("/:id", getAccount)
		api.POST("", createAccount)
		api.PUT("/:id", updateAccount)
		api.DELETE("/:id", deleteAccount)
		api.GET("/:id/balance", getBalance)
	}

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Account service starting on port %s", port)
	router.Run(":" + port)
}

// Temporary handlers (will implement with DB later)
func listAccounts(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "List all accounts - to be implemented",
		"data":    []interface{}{},
	})
}

func getAccount(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{
		"message": "Get account by ID - to be implemented",
		"id":      id,
	})
}

func createAccount(c *gin.Context) {
	c.JSON(http.StatusCreated, gin.H{
		"message": "Create account - to be implemented",
	})
}

func updateAccount(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{
		"message": "Update account - to be implemented",
		"id":      id,
	})
}

func deleteAccount(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{
		"message": "Delete account - to be implemented",
		"id":      id,
	})
}

func getBalance(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{
		"message": "Get account balance - to be implemented",
		"id":      id,
		"balance": 0,
	})
}
