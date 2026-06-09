package middleware

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"

	pb "github.com/video-converter/shared/proto/gen/go/shared/proto"
	"github.com/video-converter/gateway/internal/clients"
	"github.com/video-converter/gateway/internal/models"
)

// CORS middleware with configurable origins
func CORS(origins string) gin.HandlerFunc {
	config := cors.DefaultConfig()
	config.AllowOrigins = strings.Split(origins, ",")
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
	config.AllowCredentials = true
	return cors.New(config)
}

// Logging middleware with enhanced monitoring
func Logger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// Color coding for different status codes
		var statusColor string
		switch {
		case param.StatusCode >= 200 && param.StatusCode < 300:
			statusColor = "\033[32m" // Green
		case param.StatusCode >= 300 && param.StatusCode < 400:
			statusColor = "\033[33m" // Yellow
		case param.StatusCode >= 400 && param.StatusCode < 500:
			statusColor = "\033[31m" // Red
		case param.StatusCode >= 500:
			statusColor = "\033[35m" // Magenta
		default:
			statusColor = "\033[0m" // Reset
		}

		return fmt.Sprintf("%s[%s] %s%d\033[0m %s %s %s %s \"%s\" %s\n",
			statusColor,
			param.TimeStamp.Format("2006/01/02 - 15:04:05"),
			statusColor,
			param.StatusCode,
			param.Method,
			param.Path,
			param.Latency,
			param.ClientIP,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	})
}

// Recovery middleware with better error handling
func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = "unknown"
		}

		log.Printf("PANIC [%s] %v", requestID, recovered)

		if _, ok := recovered.(string); ok {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "Internal server error",
				Code:    "INTERNAL_ERROR",
				Details: fmt.Sprintf("Request ID: %s", requestID),
			})
		} else {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "Internal server error",
				Code:    "INTERNAL_ERROR",
				Details: fmt.Sprintf("Request ID: %s", requestID),
			})
		}
		c.AbortWithStatus(http.StatusInternalServerError)
	})
}

// Request ID middleware for tracing
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			// Generate a simple request ID
			requestID = fmt.Sprintf("%d-%d", time.Now().UnixNano(), rand.Intn(1000))
		}
		c.Header("X-Request-ID", requestID)
		c.Set("requestID", requestID)
		c.Next()
	}
}

// Rate limiting middleware
type RateLimiter struct {
	visitors map[string]*rate.Limiter
	mu       sync.RWMutex
	rate     rate.Limit
	burst    int
}

func NewRateLimiter(requestsPerMinute, burst int) *RateLimiter {
	return &RateLimiter{
		visitors: make(map[string]*rate.Limiter),
		rate:     rate.Limit(requestsPerMinute) / 60, // Convert to per-second
		burst:    burst,
	}
}

func (rl *RateLimiter) getLimiter(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.visitors[ip]
	if !exists {
		limiter = rate.NewLimiter(rl.rate, rl.burst)
		rl.visitors[ip] = limiter
	}

	return limiter
}

func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		limiter := rl.getLimiter(c.ClientIP())
		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, models.ErrorResponse{
				Error: "Rate limit exceeded",
				Code:  "RATE_LIMIT_EXCEEDED",
				Details: "Too many requests, please try again later",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// Cleanup old visitors periodically
func (rl *RateLimiter) CleanupVisitors() {
	ticker := time.NewTicker(time.Minute)
	go func() {
		for {
			<-ticker.C
			rl.mu.Lock()
			for ip, limiter := range rl.visitors {
				if limiter.TokensAt(time.Now()) == float64(rl.burst) {
					delete(rl.visitors, ip)
				}
			}
			rl.mu.Unlock()
		}
	}()
}

// validateTokenViaGRPC calls the auth service to validate a JWT token
func validateTokenViaGRPC(c *gin.Context, grpcClients *clients.GRPCClients, token string) bool {
	if grpcClients == nil || grpcClients.Auth == nil {
		c.JSON(http.StatusServiceUnavailable, models.ErrorResponse{
			Error: "Authentication service unavailable",
			Code:  "AUTH_SERVICE_UNAVAILABLE",
		})
		c.Abort()
		return false
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	resp, err := grpcClients.Auth.ValidateToken(ctx, &pb.TokenRequest{Token: token})
	if err != nil {
		log.Printf("gRPC token validation error: %v", err)
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "Token validation failed",
			Code:  "TOKEN_VALIDATION_FAILED",
		})
		c.Abort()
		return false
	}

	if !resp.Valid {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "Invalid or expired token",
			Code:  "INVALID_TOKEN",
		})
		c.Abort()
		return false
	}

	c.Set("userID", resp.UserId)
	c.Set("userEmail", resp.Email)
	return true
}

// JWT Authentication middleware
func JWTAuth(grpcClients *clients.GRPCClients) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error: "Authorization header required",
				Code:  "MISSING_AUTH_HEADER",
			})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error: "Invalid authorization header format",
				Code:  "INVALID_AUTH_HEADER",
			})
			c.Abort()
			return
		}

		if !validateTokenViaGRPC(c, grpcClients, parts[1]) {
			return
		}

		c.Next()
	}
}

// Optional JWT middleware (doesn't abort on missing token)
func OptionalJWTAuth(grpcClients *clients.GRPCClients) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.Next()
			return
		}

		// Try to validate, but don't abort on failure for optional auth
		if grpcClients == nil || grpcClients.Auth == nil {
			c.Next()
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		resp, err := grpcClients.Auth.ValidateToken(ctx, &pb.TokenRequest{Token: parts[1]})
		if err == nil && resp.Valid {
			c.Set("userID", resp.UserId)
			c.Set("userEmail", resp.Email)
		}

		c.Next()
	}
}

// Security headers middleware
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Content-Security-Policy", "default-src 'self'")
		c.Next()
	}
}