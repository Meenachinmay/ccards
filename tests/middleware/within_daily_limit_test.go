package middleware

import (
	"ccards/internal/api/request"
	"ccards/pkg/middleware"
	"ccards/pkg/models"
	"ccards/tests/setup"
	"database/sql"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithinDailyLimit(t *testing.T) {
	helper := setup.NewTestHelper(t)
	db := helper.DB

	gin.SetMode(gin.TestMode)

	t.Run("within_daily_limit", func(t *testing.T) {
		cardID := uuid.New()
		companyID := uuid.New()

		insertCard(t, db, cardID, companyID)

		txReq := request.Transaction{
			CompanyID: companyID,
			CardID:    cardID,
			Amount:    100.0,
		}

		w, c := setupTestContext(txReq, cardID, companyID)

		card := getTestCard(cardID, companyID)
		c.Set("card", card)
		c.Set("transaction_amount", txReq.Amount)

		middleware.WithinDailyLimit(db)(c)

		assert.False(t, c.IsAborted())
		assert.Equal(t, http.StatusOK, w.Code)

		todaySpending, exists := c.Get("today_spending")
		assert.True(t, exists)
		assert.Equal(t, 0.0, todaySpending.(float64))

		remainingDailyLimit, exists := c.Get("remaining_daily_limit")
		assert.True(t, exists)
		assert.Equal(t, *card.DailyLimit, remainingDailyLimit.(float64))
	})

	t.Run("exceeding_daily_limit", func(t *testing.T) {
		cardID := uuid.New()
		companyID := uuid.New()

		insertCard(t, db, cardID, companyID)

		insertTransaction(t, db, cardID, 450.0)

		txReq := request.Transaction{
			CompanyID: companyID,
			CardID:    cardID,
			Amount:    100.0,
		}

		w, c := setupTestContext(txReq, cardID, companyID)

		card := getTestCard(cardID, companyID)
		c.Set("card", card)
		c.Set("transaction_amount", txReq.Amount)

		middleware.WithinDailyLimit(db)(c)

		assert.True(t, c.IsAborted())
		assert.Equal(t, http.StatusForbidden, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response["error"], "exceed daily limit")
		assert.Equal(t, 450.0, response["current_spending"])
		assert.Equal(t, 100.0, response["transaction_amount"])
		assert.Equal(t, 550.0, response["total_would_be"])
		assert.Equal(t, 50.0, response["remaining_limit"])
	})

	t.Run("within_monthly_limit", func(t *testing.T) {
		cardID := uuid.New()
		companyID := uuid.New()

		insertCard(t, db, cardID, companyID)

		txReq := request.Transaction{
			CompanyID: companyID,
			CardID:    cardID,
			Amount:    100.0,
		}

		w, c := setupTestContext(txReq, cardID, companyID)

		card := getTestCard(cardID, companyID)
		c.Set("card", card)
		c.Set("transaction_amount", txReq.Amount)

		middleware.WithinDailyLimit(db)(c)

		assert.False(t, c.IsAborted())
		assert.Equal(t, http.StatusOK, w.Code)

		monthlySpending, exists := c.Get("monthly_spending")
		assert.True(t, exists)
		assert.Equal(t, 0.0, monthlySpending.(float64))

		remainingMonthlyLimit, exists := c.Get("remaining_monthly_limit")
		assert.True(t, exists)
		assert.Equal(t, *card.MonthlyLimit, remainingMonthlyLimit.(float64))
	})

	t.Run("exceeding_monthly_limit", func(t *testing.T) {
		cardID := uuid.New()
		companyID := uuid.New()

		insertCard(t, db, cardID, companyID)

		insertTransactionWithDate(t, db, cardID, 4900.0, time.Now().AddDate(0, 0, -1)) // Monthly limit is 5000.0

		txReq := request.Transaction{
			CompanyID: companyID,
			CardID:    cardID,
			Amount:    200.0,
		}

		w, c := setupTestContext(txReq, cardID, companyID)

		card := getTestCard(cardID, companyID)
		dailyLimit := 10000.0
		card.DailyLimit = &dailyLimit
		c.Set("card", card)
		c.Set("transaction_amount", txReq.Amount)

		middleware.WithinDailyLimit(db)(c)

		assert.True(t, c.IsAborted())
		assert.Equal(t, http.StatusForbidden, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response["error"], "exceed monthly limit")
		assert.Equal(t, 4900.0, response["current_spending"])
		assert.Equal(t, 200.0, response["transaction_amount"])
		assert.Equal(t, 5100.0, response["total_would_be"])
		assert.Equal(t, 100.0, response["remaining_limit"])
	})

	t.Run("no_daily_limit", func(t *testing.T) {
		cardID := uuid.New()
		companyID := uuid.New()

		insertCard(t, db, cardID, companyID)

		txReq := request.Transaction{
			CompanyID: companyID,
			CardID:    cardID,
			Amount:    1000.0,
		}

		w, c := setupTestContext(txReq, cardID, companyID)

		card := getTestCard(cardID, companyID)
		card.DailyLimit = nil
		card.MonthlyLimit = nil
		c.Set("card", card)
		c.Set("transaction_amount", txReq.Amount)

		middleware.WithinDailyLimit(db)(c)

		assert.False(t, c.IsAborted())
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("missing_card_in_context", func(t *testing.T) {
		cardID := uuid.New()
		companyID := uuid.New()

		txReq := request.Transaction{
			CompanyID: companyID,
			CardID:    cardID,
			Amount:    100.0,
		}

		w, c := setupTestContext(txReq, cardID, companyID)

		c.Set("transaction_amount", txReq.Amount)

		middleware.WithinDailyLimit(db)(c)

		assert.True(t, c.IsAborted())
		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response["error"], "Card information not found")
	})

	t.Run("invalid_card_type_in_context", func(t *testing.T) {
		cardID := uuid.New()
		companyID := uuid.New()

		txReq := request.Transaction{
			CompanyID: companyID,
			CardID:    cardID,
			Amount:    100.0,
		}

		w, c := setupTestContext(txReq, cardID, companyID)

		c.Set("card", "not a card object")
		c.Set("transaction_amount", txReq.Amount)

		middleware.WithinDailyLimit(db)(c)

		assert.True(t, c.IsAborted())
		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response["error"], "Invalid card information")
	})

	t.Run("missing_transaction_amount_in_context", func(t *testing.T) {
		cardID := uuid.New()
		companyID := uuid.New()

		txReq := request.Transaction{
			CompanyID: companyID,
			CardID:    cardID,
			Amount:    100.0,
		}

		w, c := setupTestContext(txReq, cardID, companyID)

		card := getTestCard(cardID, companyID)
		c.Set("card", card)

		middleware.WithinDailyLimit(db)(c)

		assert.True(t, c.IsAborted())
		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response["error"], "Transaction amount not found")
	})

	t.Run("invalid_transaction_amount_type", func(t *testing.T) {
		cardID := uuid.New()
		companyID := uuid.New()

		txReq := request.Transaction{
			CompanyID: companyID,
			CardID:    cardID,
			Amount:    100.0,
		}

		w, c := setupTestContext(txReq, cardID, companyID)

		card := getTestCard(cardID, companyID)
		c.Set("card", card)
		c.Set("transaction_amount", "not a number")

		middleware.WithinDailyLimit(db)(c)

		assert.True(t, c.IsAborted())
		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response["error"], "Invalid transaction amount")
	})
}

func insertTransaction(t *testing.T, db *sql.DB, cardID uuid.UUID, amount float64) {
	insertTransactionWithDate(t, db, cardID, amount, time.Now())
}

func insertTransactionWithDate(t *testing.T, db *sql.DB, cardID uuid.UUID, amount float64, transactionDate time.Time) {
	var companyID uuid.UUID
	err := db.QueryRow("SELECT company_id FROM cards WHERE id = $1", cardID).Scan(&companyID)
	require.NoError(t, err)

	query := `
		INSERT INTO transactions (
			id, card_id, company_id, amount, status, transaction_type, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8
		)
	`

	_, err = db.Exec(
		query,
		uuid.New(),
		cardID,
		companyID,
		amount,
		models.TransactionStatusCompleted,
		models.TransactionTypePurchase,
		transactionDate,
		transactionDate,
	)
	require.NoError(t, err)
}
