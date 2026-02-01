package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// JWT secret key - in production, use environment variable
var jwtSecret = []byte(getEnv("JWT_SECRET", "your-256-bit-secret-key-change-in-production"))

// Claims represents the JWT claims structure
type Claims struct {
	UserID int64  `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// LoginRequest represents the login request body
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse represents the login response from user service
type LoginResponse struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

func main() {
	// Create Gin router
	router := gin.Default()

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "api-gateway",
		})
	})

	// Public routes (no authentication needed)
	public := router.Group("/api/v1")
	{
		public.POST("/auth/login", handleLogin)
	}

	// Protected routes (require JWT authentication)
	protected := router.Group("/api/v1")
	protected.Use(authMiddleware())
	{
		// Proxy to User Service
		protected.Any("/users", proxyToUserService)
		protected.Any("/users/*path", proxyToUserService)

		// Proxy to Account Service
		protected.Any("/accounts", proxyToAccountService)
		protected.Any("/accounts/*path", proxyToAccountService)

		// Proxy to Payment Service
		protected.Any("/payments", proxyToPaymentService)
		protected.Any("/payments/*path", proxyToPaymentService)

		// Proxy to Transfer Service
		protected.Any("/transfers", proxyToTransferService)
		protected.Any("/transfers/*path", proxyToTransferService)

		// Proxy to Notification Service
		protected.Any("/notifications", proxyToNotificationService)
		protected.Any("/notifications/*path", proxyToNotificationService)
	}

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("API Gateway starting on port %s", port)
	router.Run(":" + port)
}

// authMiddleware validates JWT tokens and extracts claims
func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header required",
			})
			return
		}

		// Extract token from "Bearer <token>" format
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid authorization header format",
			})
			return
		}

		// Parse and validate token
		token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return jwtSecret, nil
		})

		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired token",
			})
			return
		}

		claims, ok := token.Claims.(*Claims)
		if !ok || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token claims",
			})
			return
		}

		// Store claims in context for use by handlers
		c.Set("user_id", claims.UserID)
		c.Set("user_role", claims.Role)

		c.Next()
	}
}

// handleLogin authenticates user and returns JWT token
func handleLogin(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	// Forward login request to user service
	userServiceURL := getEnv("USER_SERVICE_URL", "http://user.user.svc.cluster.local:8080")
	loginURL := userServiceURL + "/api/users/login"

	reqBody, _ := json.Marshal(req)
	resp, err := http.Post(loginURL, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		log.Printf("Error contacting user service: %v", err)
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Authentication service unavailable",
		})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid credentials",
		})
		return
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("User service returned error: %d - %s", resp.StatusCode, string(body))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Authentication failed",
		})
		return
	}

	var loginResp LoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		log.Printf("Error decoding user service response: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Authentication failed",
		})
		return
	}

	// Generate JWT token
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		UserID: loginResp.ID,
		Role:   loginResp.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   loginResp.Username,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		log.Printf("Error signing token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate token",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":      tokenString,
		"expires_at": expirationTime.Format(time.RFC3339),
		"user": gin.H{
			"id":       loginResp.ID,
			"username": loginResp.Username,
			"role":     loginResp.Role,
		},
	})
}

// Proxy handlers for each service
func proxyToUserService(c *gin.Context) {
	userServiceURL := getEnv("USER_SERVICE_URL", "http://user.user.svc.cluster.local:8080")
	proxyRequest(c, userServiceURL, "/api/v1/users", "/api/users")
}

func proxyToAccountService(c *gin.Context) {
	accountServiceURL := getEnv("ACCOUNT_SERVICE_URL", "http://account.account.svc.cluster.local:8080")
	proxyRequest(c, accountServiceURL, "/api/v1/accounts", "/api/accounts")
}

func proxyToPaymentService(c *gin.Context) {
	paymentServiceURL := getEnv("PAYMENT_SERVICE_URL", "http://payment.payment.svc.cluster.local:8080")
	proxyRequest(c, paymentServiceURL, "/api/v1/payments", "/api/payments")
}

func proxyToTransferService(c *gin.Context) {
	transferServiceURL := getEnv("TRANSFER_SERVICE_URL", "http://transfer.transfer.svc.cluster.local:8080")
	proxyRequest(c, transferServiceURL, "/api/v1/transfers", "/api/transfers")
}

func proxyToNotificationService(c *gin.Context) {
	notificationServiceURL := getEnv("NOTIFICATION_SERVICE_URL", "http://notification.notification.svc.cluster.local:8080")
	proxyRequest(c, notificationServiceURL, "/api/v1/notifications", "/api/notifications")
}

// proxyRequest forwards the request to the target service with user context headers
func proxyRequest(c *gin.Context, targetURL, stripPrefix, addPrefix string) {
	target, err := url.Parse(targetURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid service URL"})
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(target)

	// Modify the request path
	originalPath := c.Request.URL.Path
	newPath := strings.TrimPrefix(originalPath, stripPrefix)
	if addPrefix != "" && !strings.HasPrefix(newPath, addPrefix) {
		newPath = addPrefix + newPath
	}

	// Get user context from middleware (if available)
	userID, _ := c.Get("user_id")
	userRole, _ := c.Get("user_role")

	// Director modifies the request before sending it to the backend
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.URL.Path = newPath
		req.Host = target.Host

		// Propagate user context to backend services
		if userID != nil {
			req.Header.Set("X-User-ID", formatInt64(userID))
		}
		if userRole != nil {
			req.Header.Set("X-User-Role", userRole.(string))
		}

		log.Printf("Proxying: %s %s -> %s%s (User: %v, Role: %v)",
			req.Method, originalPath, target.Host, newPath, userID, userRole)
	}

	// Handle errors
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("Proxy error: %v", err)
		c.JSON(http.StatusBadGateway, gin.H{
			"error":   "Service unavailable",
			"service": target.Host,
		})
	}

	proxy.ServeHTTP(c.Writer, c.Request)
}

func formatInt64(v interface{}) string {
	if i, ok := v.(int64); ok {
		return strconv.FormatInt(i, 10)
	}
	return ""
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
