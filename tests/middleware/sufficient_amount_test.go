package middleware

import (
	"ccards/internal/api/request"
	"ccards/pkg/middleware"
	"ccards/pkg/models"
	"ccards/tests/setup"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSufficientAmount(t *testing.T) {
	helper := setup.NewTestHelper(t)
	db := helper.DB

	_ = models.CardStatusActive

	gin.SetMode(gin.TestMode)

	t.Run("sufficient_balance_and_within_spending_limit", func(t *testing.T) {
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
		c.Set("transaction_request", &txReq)

		middleware.SufficientAmount()(c)

		assert.False(t, c.IsAborted())
		assert.Equal(t, http.StatusOK, w.Code)

		transactionAmount, exists := c.Get("transaction_amount")
		assert.True(t, exists)
		assert.Equal(t, txReq.Amount, transactionAmount.(float64))
	})

	t.Run("insufficient_balance", func(t *testing.T) {
		cardID := uuid.New()
		companyID := uuid.New()

		insertCard(t, db, cardID, companyID)

		txReq := request.Transaction{
			CompanyID: companyID,
			CardID:    cardID,
			Amount:    3000.0,
		}

		w, c := setupTestContext(txReq, cardID, companyID)

		card := getTestCard(cardID, companyID)
		c.Set("card", card)
		c.Set("transaction_request", &txReq)

		middleware.SufficientAmount()(c)

		assert.True(t, c.IsAborted())
		assert.Equal(t, http.StatusPaymentRequired, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response["error"], "Insufficient balance")
		assert.Equal(t, 2000.0, response["available_balance"])
		assert.Equal(t, 3000.0, response["required_amount"])
		assert.Equal(t, 1000.0, response["shortage"])
	})

	t.Run("exceeding_spending_limit", func(t *testing.T) {
		cardID := uuid.New()
		companyID := uuid.New()

		insertCard(t, db, cardID, companyID)

		txReq := request.Transaction{
			CompanyID: companyID,
			CardID:    cardID,
			Amount:    1500.0,
		}

		w, c := setupTestContext(txReq, cardID, companyID)

		card := getTestCard(cardID, companyID)
		c.Set("card", card)
		c.Set("transaction_request", &txReq)

		middleware.SufficientAmount()(c)

		assert.True(t, c.IsAborted())
		assert.Equal(t, http.StatusForbidden, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response["error"], "Transaction exceeds spending limit")
		assert.Equal(t, 1000.0, response["spending_limit"])
		assert.Equal(t, 1500.0, response["amount"])
	})

	t.Run("no_spending_limit", func(t *testing.T) {
		cardID := uuid.New()
		companyID := uuid.New()

		insertCard(t, db, cardID, companyID)

		txReq := request.Transaction{
			CompanyID: companyID,
			CardID:    cardID,
			Amount:    1500.0,
		}

		w, c := setupTestContext(txReq, cardID, companyID)

		card := getTestCard(cardID, companyID)
		card.SpendingLimit = nil
		c.Set("card", card)
		c.Set("transaction_request", &txReq)

		middleware.SufficientAmount()(c)

		assert.False(t, c.IsAborted())
		assert.Equal(t, http.StatusOK, w.Code)

		transactionAmount, exists := c.Get("transaction_amount")
		assert.True(t, exists)
		assert.Equal(t, txReq.Amount, transactionAmount.(float64))
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

		c.Set("transaction_request", &txReq)

		middleware.SufficientAmount()(c)

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
		c.Set("transaction_request", &txReq)

		middleware.SufficientAmount()(c)

		assert.True(t, c.IsAborted())
		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response["error"], "Invalid card information")
	})

	t.Run("missing_transaction_request_in_context", func(t *testing.T) {
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

		middleware.SufficientAmount()(c)

		assert.True(t, c.IsAborted())
		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response["error"], "Transaction request not found")
	})

	t.Run("invalid_transaction_request_type", func(t *testing.T) {
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
		c.Set("transaction_request", "not a transaction request")

		middleware.SufficientAmount()(c)

		assert.True(t, c.IsAborted())
		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response["error"], "Invalid transaction request")
	})
}
