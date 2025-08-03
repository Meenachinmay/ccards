package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"ccards/pkg/config"
	"ccards/pkg/utils"
)

type ClientAuthMiddleware struct {
	jwtConfig   config.JWTConfig
	redisClient *redis.Client
}

func NewClientAuthMiddleware(cfg *config.Config, redisClient *redis.Client) *ClientAuthMiddleware {
	return &ClientAuthMiddleware{
		jwtConfig:   cfg.JWT,
		redisClient: redisClient,
	}
}

func (m *ClientAuthMiddleware) ClientAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header is required",
			})
			c.Abort()
			return
		}

		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid authorization header format",
			})
			c.Abort()
			return
		}

		token := tokenParts[1]

		claims, err := utils.ValidateJWT(token, m.jwtConfig.Secret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired token",
			})
			c.Abort()
			return
		}

		sessionKey := fmt.Sprintf("session:%s", claims.CompanyID.String())
		exists, err := m.redisClient.Exists(context.Background(), sessionKey).Result()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Session validation failed",
			})
			c.Abort()
			return
		}

		if exists == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Session expired or invalid",
			})
			c.Abort()
			return
		}

		sessionData, err := m.redisClient.HGetAll(context.Background(), sessionKey).Result()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to retrieve session data",
			})
			c.Abort()
			return
		}

		if sessionData["company_id"] != claims.CompanyID.String() {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Session mismatch",
			})
			c.Abort()
			return
		}

		c.Set("company_id", claims.CompanyID)
		c.Set("company_id_string", claims.CompanyID.String())

		c.Next()
	}
}

func GetCompanyIDFromContext(c *gin.Context) (uuid.UUID, error) {
	companyID, exists := c.Get("company_id")
	if !exists {
		return uuid.Nil, fmt.Errorf("company ID not found in context")
	}

	id, ok := companyID.(uuid.UUID)
	if !ok {
		return uuid.Nil, fmt.Errorf("invalid company ID type in context")
	}

	return id, nil
}
