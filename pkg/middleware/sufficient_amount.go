package middleware

import (
	"ccards/internal/api/request"
	"net/http"

	"github.com/gin-gonic/gin"

	"ccards/pkg/models"
)

func SufficientAmount() gin.HandlerFunc {
	return func(c *gin.Context) {
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

		reqInterface, exists := c.Get("transaction_request")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Transaction request not found",
			})
			c.Abort()
			return
		}

		reqPtr, ok := reqInterface.(*request.Transaction)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Invalid transaction request",
			})
			c.Abort()
			return
		}

		transactionAmount := reqPtr.Amount

		if card.Balance < transactionAmount {
			c.JSON(http.StatusPaymentRequired, gin.H{
				"error":             "Insufficient balance",
				"available_balance": card.Balance,
				"required_amount":   transactionAmount,
				"shortage":          transactionAmount - card.Balance,
			})
			c.Abort()
			return
		}

		if card.SpendingLimit != nil && transactionAmount > *card.SpendingLimit {
			c.JSON(http.StatusForbidden, gin.H{
				"error":          "Transaction exceeds spending limit",
				"spending_limit": *card.SpendingLimit,
				"amount":         transactionAmount,
			})
			c.Abort()
			return
		}

		c.Set("transaction_amount", transactionAmount)

		c.Next()
	}
}
