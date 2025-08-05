package middleware

import (
	"bytes"
	"ccards/internal/api/request"
	"ccards/internal/client"
	"ccards/pkg/models"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func setupTestContext(txReq request.Transaction, cardID, companyID uuid.UUID) (*httptest.ResponseRecorder, *gin.Context) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody, _ := json.Marshal(txReq)
	req, _ := http.NewRequest("POST", "/api/cards/transactions", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	c.Request = req
	c.Params = gin.Params{
		{Key: "card_id", Value: cardID.String()},
	}

	c.Request.Body = io.NopCloser(bytes.NewBuffer(reqBody))

	return w, c
}

func insertCard(t *testing.T, db *sql.DB, cardID, companyID uuid.UUID) {
	clientRepo := client.NewRepository(db)
	ctx := context.Background()

	company := &models.Company{
		ID:       companyID,
		ClientID: uuid.New(),
		Name:     "Test Company",
		Email:    fmt.Sprintf("test-%s@example.com", companyID.String()[:8]),
		Password: "hashed_password",
		Address:  "123 Test St",
		Phone:    "123-456-7890",
		Status:   models.CompanyStatusActive,
	}

	// Retry creating company with backoff
	maxRetries := 5
	var err error
	for i := 0; i < maxRetries; i++ {
		err = clientRepo.CreateCompany(ctx, company)
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
	require.NoError(t, err, "Failed to create company after %d retries", maxRetries)

	spendingLimit := 1000.0
	dailyLimit := 500.0
	monthlyLimit := 5000.0

	cardNumber := fmt.Sprintf("4111%s", cardID.String()[:12])
	lastFour := cardNumber[len(cardNumber)-4:]

	card := &models.Card{
		ID:             cardID,
		CompanyID:      companyID,
		CardNumber:     cardNumber,
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
		CVVHash:        "test-cvv-hash",
		LastFour:       lastFour,
	}

	cards := []*models.Card{card}

	// Retry creating cards with backoff
	for i := 0; i < maxRetries; i++ {
		err = clientRepo.CreateCardsInBatch(ctx, cards)
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
	require.NoError(t, err, "Failed to create cards after %d retries", maxRetries)
}
