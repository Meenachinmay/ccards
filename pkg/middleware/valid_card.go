package middleware

import (
	"ccards/internal/api/request"
	"ccards/pkg/models"
	"database/sql"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func ValidCard(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		companyIDInterface, exists := c.Get("company_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized access"})
			c.Abort()
			return
		}

		companyID, ok := companyIDInterface.(uuid.UUID)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid company ID format"})
			c.Abort()
			return
		}

		var txReq request.Transaction
		if err := c.ShouldBindJSON(&txReq); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			c.Abort()
			return
		}

		if txReq.CompanyID != companyID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Company ID mismatch"})
			c.Abort()
			return
		}

		// Fetch card from the database
		query := `SELECT id, company_id, card_number, card_holder_name, employee_id, employee_email, card_type, status, balance, spending_limit, daily_limit, monthly_limit, expiry_date, cvv_hash, last_four, created_at, updated_at, blocked_at, blocked_reason FROM cards WHERE id = $1`
		row := db.QueryRowContext(c, query, txReq.CardID)

		var card models.Card
		err := row.Scan(
			&card.ID, &card.CompanyID, &card.CardNumber, &card.CardHolderName, &card.EmployeeID, &card.EmployeeEmail,
			&card.CardType, &card.Status, &card.Balance, &card.SpendingLimit, &card.DailyLimit, &card.MonthlyLimit,
			&card.ExpiryDate, &card.CVVHash, &card.LastFour, &card.CreatedAt, &card.UpdatedAt, &card.BlockedAt, &card.BlockedReason,
		)

		if err != nil {
			if errors.Is(sql.ErrNoRows, err) {
				c.JSON(http.StatusNotFound, gin.H{"error": "Card not found"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			}
			c.Abort()
			return
		}

		c.Set("card", &card)
		c.Set("transaction_request", &txReq)
		c.Next()
	}
}
