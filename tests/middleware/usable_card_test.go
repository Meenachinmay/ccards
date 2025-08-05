package middleware

import (
	"ccards/internal/api/request"
	"ccards/pkg/middleware"
	"ccards/pkg/models"
	"ccards/tests/setup"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUsableCard(t *testing.T) {
	helper := setup.NewTestHelper(t)
	db := helper.DB

	gin.SetMode(gin.TestMode)

	t.Run("active_card", func(t *testing.T) {
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

		middleware.UsableCard()(c)

		assert.False(t, c.IsAborted())
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("blocked_card", func(t *testing.T) {
		cardID := uuid.New()
		companyID := uuid.New()

		txReq := request.Transaction{
			CompanyID: companyID,
			CardID:    cardID,
			Amount:    100.0,
		}

		w, c := setupTestContext(txReq, cardID, companyID)

		card := getTestCard(cardID, companyID)
		card.Status = models.CardStatusBlocked
		blockedReason := "Suspicious activity"
		card.BlockedReason = &blockedReason
		c.Set("card", card)

		middleware.UsableCard()(c)

		assert.True(t, c.IsAborted())
		assert.Equal(t, http.StatusForbidden, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response["error"], "Card is blocked")
		assert.Contains(t, response["error"], "Suspicious activity")
		assert.Equal(t, "blocked", response["status"])
	})

	t.Run("expired_card", func(t *testing.T) {
		cardID := uuid.New()
		companyID := uuid.New()

		txReq := request.Transaction{
			CompanyID: companyID,
			CardID:    cardID,
			Amount:    100.0,
		}

		w, c := setupTestContext(txReq, cardID, companyID)

		card := getTestCard(cardID, companyID)
		card.ExpiryDate = time.Now().AddDate(-1, 0, 0)
		c.Set("card", card)

		middleware.UsableCard()(c)

		assert.True(t, c.IsAborted())
		assert.Equal(t, http.StatusForbidden, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response["error"], "Card has expired")
		assert.Contains(t, response, "expiry_date")
	})

	t.Run("cancelled_card", func(t *testing.T) {
		cardID := uuid.New()
		companyID := uuid.New()

		txReq := request.Transaction{
			CompanyID: companyID,
			CardID:    cardID,
			Amount:    100.0,
		}

		w, c := setupTestContext(txReq, cardID, companyID)

		card := getTestCard(cardID, companyID)
		card.Status = models.CardStatusCancelled
		c.Set("card", card)

		middleware.UsableCard()(c)

		assert.True(t, c.IsAborted())
		assert.Equal(t, http.StatusForbidden, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response["error"], "Card has been cancelled")
		assert.Equal(t, "cancelled", response["status"])
	})

	t.Run("card_expiring_soon", func(t *testing.T) {
		cardID := uuid.New()
		companyID := uuid.New()

		txReq := request.Transaction{
			CompanyID: companyID,
			CardID:    cardID,
			Amount:    100.0,
		}

		w, c := setupTestContext(txReq, cardID, companyID)

		card := getTestCard(cardID, companyID)
		card.ExpiryDate = time.Now().AddDate(0, 0, 15)
		c.Set("card", card)

		middleware.UsableCard()(c)

		assert.False(t, c.IsAborted())
		assert.Equal(t, http.StatusOK, w.Code)

		expiryWarning, exists := c.Get("expiry_warning")
		assert.True(t, exists)
		assert.True(t, expiryWarning.(bool))

		daysUntilExpiry, exists := c.Get("days_until_expiry")
		assert.True(t, exists)
		assert.LessOrEqual(t, daysUntilExpiry.(int), 15)
		assert.GreaterOrEqual(t, daysUntilExpiry.(int), 14)
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

		middleware.UsableCard()(c)

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

		middleware.UsableCard()(c)

		assert.True(t, c.IsAborted())
		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response["error"], "Invalid card information")
	})
}
