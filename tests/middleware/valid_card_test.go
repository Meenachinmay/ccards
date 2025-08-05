package middleware

import (
	"bytes"
	"ccards/internal/api/request"
	"ccards/pkg/middleware"
	"ccards/pkg/models"
	"ccards/tests/setup"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidCard(t *testing.T) {
	helper := setup.NewTestHelper(t)
	db := helper.DB

	gin.SetMode(gin.TestMode)

	t.Run("valid_card", func(t *testing.T) {
		cardID := uuid.New()
		companyID := uuid.New()

		insertCard(t, db, cardID, companyID)

		txReq := request.Transaction{
			CompanyID:        companyID,
			CardID:           cardID,
			Amount:           100.0,
			MerchantCategory: "test",
		}

		w, c := setupTestContextWithCompanyID(txReq, cardID, companyID)

		middleware.ValidCard(db)(c)
		// Add this debugging code
		if w.Code != http.StatusOK {
			t.Logf("Response body: %s", w.Body.String())
		}

		assert.False(t, c.IsAborted())
		assert.Equal(t, http.StatusOK, w.Code)

		cardInterface, exists := c.Get("card")
		assert.True(t, exists)
		card, ok := cardInterface.(*models.Card)
		assert.True(t, ok)
		assert.Equal(t, cardID, card.ID)
		assert.Equal(t, companyID, card.CompanyID)

		txInterface, exists := c.Get("transaction_request")
		assert.True(t, exists)
		tx, ok := txInterface.(*request.Transaction)
		assert.True(t, ok)
		assert.Equal(t, cardID, tx.CardID)
		assert.Equal(t, companyID, tx.CompanyID)
	})

	t.Run("company_id_mismatch", func(t *testing.T) {
		cardID := uuid.New()
		companyID := uuid.New()
		wrongCompanyID := uuid.New()

		insertCard(t, db, cardID, companyID)

		txReq := request.Transaction{
			CompanyID:        wrongCompanyID,
			CardID:           cardID,
			Amount:           100.0,
			MerchantCategory: "test",
		}

		w, c := setupTestContextWithCompanyID(txReq, cardID, companyID)

		middleware.ValidCard(db)(c)

		assert.True(t, c.IsAborted())
		assert.Equal(t, http.StatusForbidden, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response["error"], "Company ID mismatch")
	})

	t.Run("card_not_found", func(t *testing.T) {
		cardID := uuid.New()
		companyID := uuid.New()

		txReq := request.Transaction{
			CompanyID:        companyID,
			CardID:           cardID,
			Amount:           100.0,
			MerchantCategory: "test",
		}

		w, c := setupTestContextWithCompanyID(txReq, cardID, companyID)

		middleware.ValidCard(db)(c)

		assert.True(t, c.IsAborted())
		assert.Equal(t, http.StatusNotFound, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response["error"], "Card not found")
	})

	t.Run("missing_company_id_in_context", func(t *testing.T) {
		cardID := uuid.New()
		companyID := uuid.New()

		txReq := request.Transaction{
			CompanyID:        companyID,
			CardID:           cardID,
			Amount:           100.0,
			MerchantCategory: "test",
		}

		w, c := setupTestContext(txReq, cardID, companyID)

		middleware.ValidCard(db)(c)

		assert.True(t, c.IsAborted())
		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response["error"], "Unauthorized access")
	})

	t.Run("invalid_request_body", func(t *testing.T) {
		cardID := uuid.New()
		companyID := uuid.New()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		req, _ := http.NewRequest("POST", "/api/cards/transactions", bytes.NewBuffer([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")

		c.Request = req
		c.Params = gin.Params{
			{Key: "card_id", Value: cardID.String()},
		}
		c.Set("company_id", companyID)

		middleware.ValidCard(db)(c)

		assert.True(t, c.IsAborted())
		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response["error"], "Invalid request body")
	})
}

func setupTestContextWithCompanyID(txReq request.Transaction, cardID, companyID uuid.UUID) (*httptest.ResponseRecorder, *gin.Context) {
	w, c := setupTestContext(txReq, cardID, companyID)
	c.Set("company_id", companyID)
	return w, c
}
