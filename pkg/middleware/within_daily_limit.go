package middleware

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"ccards/pkg/models"
)

type DailyLimitMiddleware struct {
	db *sql.DB
}

func NewDailyLimitMiddleware(db *sql.DB) *DailyLimitMiddleware {
	return &DailyLimitMiddleware{
		db: db,
	}
}

func WithinDailyLimit(db *sql.DB) gin.HandlerFunc {
	m := NewDailyLimitMiddleware(db)
	return m.Handle()
}

func (m *DailyLimitMiddleware) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get card from context
		cardInterface, exists := c.Get("card")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Card information not found",
			})
			c.Abort()
			return
		}

		card, ok := cardInterface.(*models.Card)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Invalid card information",
			})
			c.Abort()
			return
		}

		if card.DailyLimit == nil {
			c.Next()
			return
		}

		amountInterface, exists := c.Get("transaction_amount")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Transaction amount not found",
			})
			c.Abort()
			return
		}

		transactionAmount, ok := amountInterface.(float64)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Invalid transaction amount",
			})
			c.Abort()
			return
		}

		todaySpending, err := m.getTodaySpending(c.Request.Context(), card.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to calculate daily spending",
			})
			c.Abort()
			return
		}

		totalDailySpending := todaySpending + transactionAmount
		if totalDailySpending > *card.DailyLimit {
			c.JSON(http.StatusForbidden, gin.H{
				"error":              "Transaction would exceed daily limit",
				"daily_limit":        *card.DailyLimit,
				"current_spending":   todaySpending,
				"transaction_amount": transactionAmount,
				"total_would_be":     totalDailySpending,
				"remaining_limit":    *card.DailyLimit - todaySpending,
			})
			c.Abort()
			return
		}

		c.Set("today_spending", todaySpending)
		c.Set("remaining_daily_limit", *card.DailyLimit-todaySpending)

		if card.MonthlyLimit != nil {
			monthlySpending, err := m.getMonthlySpending(c.Request.Context(), card.ID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Failed to calculate monthly spending",
				})
				c.Abort()
				return
			}

			totalMonthlySpending := monthlySpending + transactionAmount
			if totalMonthlySpending > *card.MonthlyLimit {
				c.JSON(http.StatusForbidden, gin.H{
					"error":              "Transaction would exceed monthly limit",
					"monthly_limit":      *card.MonthlyLimit,
					"current_spending":   monthlySpending,
					"transaction_amount": transactionAmount,
					"total_would_be":     totalMonthlySpending,
					"remaining_limit":    *card.MonthlyLimit - monthlySpending,
				})
				c.Abort()
				return
			}

			c.Set("monthly_spending", monthlySpending)
			c.Set("remaining_monthly_limit", *card.MonthlyLimit-monthlySpending)
		}

		c.Next()
	}
}

func (m *DailyLimitMiddleware) getTodaySpending(ctx context.Context, cardID uuid.UUID) (float64, error) {
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	query := `
		SELECT COALESCE(SUM(amount), 0) as total_spending
		FROM transactions
		WHERE card_id = $1 
		  AND status = $2
		  AND created_at >= $3
		  AND transaction_type = $4
	`

	var totalSpending float64
	err := m.db.QueryRowContext(ctx, query,
		cardID,
		models.TransactionStatusCompleted,
		startOfDay,
		models.TransactionTypePurchase,
	).Scan(&totalSpending)

	if err != nil {
		return 0, err
	}

	return totalSpending, nil
}

func (m *DailyLimitMiddleware) getMonthlySpending(ctx context.Context, cardID uuid.UUID) (float64, error) {
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	query := `
		SELECT COALESCE(SUM(amount), 0) as total_spending
		FROM transactions
		WHERE card_id = $1 
		  AND status = $2
		  AND created_at >= $3
		  AND transaction_type = $4
	`

	var totalSpending float64
	err := m.db.QueryRowContext(ctx, query,
		cardID,
		models.TransactionStatusCompleted,
		startOfMonth,
		models.TransactionTypePurchase,
	).Scan(&totalSpending)

	if err != nil {
		return 0, err
	}

	return totalSpending, nil
}
