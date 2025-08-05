package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"ccards/pkg/models"
)

func UsableCard() gin.HandlerFunc {
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

		if card.Status != models.CardStatusActive {
			var errorMessage string
			switch card.Status {
			case models.CardStatusBlocked:
				errorMessage = "Card is blocked"
				if card.BlockedReason != nil {
					errorMessage += ": " + *card.BlockedReason
				}
			case models.CardStatusExpired:
				errorMessage = "Card has expired"
			case models.CardStatusCancelled:
				errorMessage = "Card has been cancelled"
			default:
				errorMessage = "Card is not active"
			}

			c.JSON(http.StatusForbidden, gin.H{
				"error":  errorMessage,
				"status": card.Status,
			})
			c.Abort()
			return
		}

		if time.Now().After(card.ExpiryDate) {
			c.JSON(http.StatusForbidden, gin.H{
				"error":       "Card has expired",
				"expiry_date": card.ExpiryDate.Format("2006-01-02"),
			})
			c.Abort()
			return
		}

		daysUntilExpiry := int(time.Until(card.ExpiryDate).Hours() / 24)
		if daysUntilExpiry <= 30 && daysUntilExpiry >= 0 {
			c.Set("expiry_warning", true)
			c.Set("days_until_expiry", daysUntilExpiry)
		}

		c.Next()
	}
}
