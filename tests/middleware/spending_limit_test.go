package middleware

import (
	"ccards/internal/api/request"
	"ccards/pkg/middleware"
	"ccards/pkg/models"
	"ccards/tests/setup"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSpendingLimit(t *testing.T) {
	helper := setup.NewTestHelper(t)
	db := helper.DB

	gin.SetMode(gin.TestMode)

	t.Run("merchant_category_allowed", func(t *testing.T) {
		cardID := uuid.New()
		companyID := uuid.New()

		insertCard(t, db, cardID, companyID)

		createMerchantCategoryControl(t, db, cardID, []string{"food"}, []string{})

		txReq := request.Transaction{
			CompanyID:        companyID,
			CardID:           cardID,
			Amount:           100.0,
			MerchantCategory: "food",
		}

		w, c := setupTestContext(txReq, cardID, companyID)

		c.Set("card", getTestCard(cardID, companyID))
		c.Set("transaction_request", &txReq)

		middleware.SpendingLimit(db)(c)

		assert.False(t, c.IsAborted())
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("merchant_category_blocked", func(t *testing.T) {
		cardID := uuid.New()
		companyID := uuid.New()

		insertCard(t, db, cardID, companyID)

		createMerchantCategoryControl(t, db, cardID, []string{"food"}, []string{})

		txReq := request.Transaction{
			CompanyID:        companyID,
			CardID:           cardID,
			Amount:           100.0,
			MerchantCategory: "entertainment",
		}

		w, c := setupTestContext(txReq, cardID, companyID)

		c.Set("card", getTestCard(cardID, companyID))
		c.Set("transaction_request", &txReq)

		middleware.SpendingLimit(db)(c)

		assert.True(t, c.IsAborted())
		assert.Equal(t, http.StatusForbidden, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response["error"], "not in allowed list")
		assert.Equal(t, "merchant_category", response["control_type"])
	})

	t.Run("merchant_category_explicitly_blocked", func(t *testing.T) {
		cardID := uuid.New()
		companyID := uuid.New()

		insertCard(t, db, cardID, companyID)

		createMerchantCategoryControl(t, db, cardID, []string{}, []string{"gambling"})

		txReq := request.Transaction{
			CompanyID:        companyID,
			CardID:           cardID,
			Amount:           100.0,
			MerchantCategory: "gambling",
		}

		w, c := setupTestContext(txReq, cardID, companyID)

		c.Set("card", getTestCard(cardID, companyID))
		c.Set("transaction_request", &txReq)

		middleware.SpendingLimit(db)(c)

		assert.True(t, c.IsAborted())
		assert.Equal(t, http.StatusForbidden, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response["error"], "is not allowed")
		assert.Equal(t, "merchant_category", response["control_type"])
	})

	t.Run("time_based_within_allowed_time", func(t *testing.T) {
		cardID := uuid.New()
		companyID := uuid.New()

		insertCard(t, db, cardID, companyID)

		loc, err := time.LoadLocation(middleware.TokyoTimezone)
		require.NoError(t, err)
		now := time.Now().In(loc)

		startHour := (now.Hour() + 23) % 24 // 1 hour before current hour
		endHour := (now.Hour() + 1) % 24    // 1 hour after current hour

		startTime := fmt.Sprintf("%02d:00", startHour)
		endTime := fmt.Sprintf("%02d:00", endHour)

		createTimeBasedControl(t, db, cardID, startTime, endTime)

		txReq := request.Transaction{
			CompanyID:        companyID,
			CardID:           cardID,
			Amount:           100.0,
			MerchantCategory: "food",
		}

		w, c := setupTestContext(txReq, cardID, companyID)

		c.Set("card", getTestCard(cardID, companyID))
		c.Set("transaction_request", &txReq)

		middleware.SpendingLimit(db)(c)

		assert.False(t, c.IsAborted())
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("time_based_outside_allowed_time", func(t *testing.T) {
		cardID := uuid.New()
		companyID := uuid.New()

		insertCard(t, db, cardID, companyID)

		loc, err := time.LoadLocation(middleware.TokyoTimezone)
		require.NoError(t, err)
		now := time.Now().In(loc)

		startHour := (now.Hour() + 2) % 24 // 2 hours after current hour
		endHour := (now.Hour() + 4) % 24   // 4 hours after current hour

		startTime := fmt.Sprintf("%02d:00", startHour)
		endTime := fmt.Sprintf("%02d:00", endHour)

		createTimeBasedControl(t, db, cardID, startTime, endTime)

		txReq := request.Transaction{
			CompanyID:        companyID,
			CardID:           cardID,
			Amount:           100.0,
			MerchantCategory: "food",
		}

		w, c := setupTestContext(txReq, cardID, companyID)

		c.Set("card", getTestCard(cardID, companyID))
		c.Set("transaction_request", &txReq)

		middleware.SpendingLimit(db)(c)

		assert.True(t, c.IsAborted())
		assert.Equal(t, http.StatusForbidden, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response["error"], "outside allowed time window")
		assert.Equal(t, "time_based", response["control_type"])
	})

	t.Run("missing_card_in_context", func(t *testing.T) {
		txReq := request.Transaction{
			CompanyID:        uuid.New(),
			CardID:           uuid.New(),
			Amount:           100.0,
			MerchantCategory: "food",
		}

		w, c := setupTestContext(txReq, txReq.CardID, txReq.CompanyID)

		c.Set("transaction_request", &txReq)

		middleware.SpendingLimit(db)(c)

		assert.True(t, c.IsAborted())
		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response["error"], "Card information not found")
	})

	t.Run("missing_transaction_request_in_context", func(t *testing.T) {
		cardID := uuid.New()
		companyID := uuid.New()

		w, c := setupTestContext(request.Transaction{}, cardID, companyID)

		c.Set("card", getTestCard(cardID, companyID))

		middleware.SpendingLimit(db)(c)

		assert.True(t, c.IsAborted())
		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response["error"], "Transaction request not found")
	})
}

func getTestCard(cardID, companyID uuid.UUID) *models.Card {
	spendingLimit := 1000.0
	dailyLimit := 500.0
	monthlyLimit := 5000.0

	return &models.Card{
		ID:             cardID,
		CompanyID:      companyID,
		CardNumber:     "4111111111111111",
		CardHolderName: "Test User",
		EmployeeID:     "EMP123",
		EmployeeEmail:  "test@example.com",
		CardType:       models.CardTypeVirtual,
		Status:         models.CardStatusActive,
		Balance:        2000.0,
		SpendingLimit:  &spendingLimit,
		DailyLimit:     &dailyLimit,
		MonthlyLimit:   &monthlyLimit,
		ExpiryDate:     time.Now().AddDate(1, 0, 0),
		LastFour:       "1111",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
}

func createMerchantCategoryControl(t *testing.T, db *sql.DB, cardID uuid.UUID, allowedCategories, blockedCategories []string) {
	control := middleware.MerchantCategoryControl{
		AllowedCategories: allowedCategories,
		BlockedCategories: blockedCategories,
	}

	controlJSON, err := json.Marshal(control)
	require.NoError(t, err)

	query := `
		INSERT INTO spending_controls (
			id, card_id, control_type, control_value, is_active
		) VALUES (
			$1, $2, $3, $4, $5
		)
	`

	// Retry with backoff
	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		_, err = db.Exec(
			query,
			uuid.New(),
			cardID,
			"merchant_category",
			controlJSON,
			true,
		)
		if err == nil {
			break
		}

		// Check if it's a deadlock error
		if err.Error() == "pq: deadlock detected" {
			// Exponential backoff: sleep for 2^i * 10ms
			backoffTime := time.Duration(1<<uint(i)) * 10 * time.Millisecond
			time.Sleep(backoffTime)
			continue
		}

		// If it's not a deadlock error, fail immediately
		require.NoError(t, err)
	}
	require.NoError(t, err, "Failed to create merchant category control after %d retries", maxRetries)
}

func createTimeBasedControl(t *testing.T, db *sql.DB, cardID uuid.UUID, startTime, endTime string) {
	control := middleware.TimeBasedControl{
		StartTime: startTime,
		EndTime:   endTime,
	}

	controlJSON, err := json.Marshal(control)
	require.NoError(t, err)

	query := `
		INSERT INTO spending_controls (
			id, card_id, control_type, control_value, is_active
		) VALUES (
			$1, $2, $3, $4, $5
		)
	`

	// Retry with backoff
	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		_, err = db.Exec(
			query,
			uuid.New(),
			cardID,
			"time_based",
			controlJSON,
			true,
		)
		if err == nil {
			break
		}

		// Check if it's a deadlock error
		if err.Error() == "pq: deadlock detected" {
			// Exponential backoff: sleep for 2^i * 10ms
			backoffTime := time.Duration(1<<uint(i)) * 10 * time.Millisecond
			time.Sleep(backoffTime)
			continue
		}

		// If it's not a deadlock error, fail immediately
		require.NoError(t, err)
	}
	require.NoError(t, err, "Failed to create time-based control after %d retries", maxRetries)
}
