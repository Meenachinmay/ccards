package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CORS is a middleware that adds CORS headers to the response
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Set CORS headers
		c.Header("Access-Control-Allow-Origin", getAllowedOrigin(origin))
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "86400") // 24 hours

		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// getAllowedOrigin returns the appropriate origin based on the request
// In production, you should maintain a whitelist of allowed origins
func getAllowedOrigin(origin string) string {
	// For development, allow all origins
	// In production, implement proper origin validation
	allowedOrigins := map[string]bool{
		"http://localhost:5173":  true,
		"http://localhost:3001":  true,
		"http://localhost:8080":  true,
		"https://yourdomain.com": true,
	}

	if allowedOrigins[origin] {
		return origin
	}

	// Default to first allowed origin or "*" for development
	return "http://localhost:3000"
}
