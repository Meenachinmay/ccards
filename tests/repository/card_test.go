package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ccards/internal/card"
	"ccards/internal/client"
	"ccards/pkg/models"
	"ccards/tests/setup"
)

func TestGetCardsByCompanyID(t *testing.T) {
	helper := setup.NewTestHelper(t)
	cardRepo := card.NewRepository(helper.DB)
	clientRepo := client.NewRepository(helper.DB)
	ctx := context.Background()

	t.Run("company_with_cards", func(t *testing.T) {
		companyID := uuid.New()
		company := &models.Company{
			ID:       companyID,
			ClientID: uuid.New(),
			Name:     "Test Company",
			Email:    "test-card-company@example.com",
			Password: "hashed_password",
			Address:  "123 Test St",
			Phone:    "123-456-7890",
			Status:   models.CompanyStatusActive,
		}
		err := clientRepo.CreateCompany(ctx, company)
		require.NoError(t, err)

		cards := []*models.Card{
			{
				ID:             uuid.New(),
				CompanyID:      companyID,
				CardNumber:     "4111111111111111",
				CardHolderName: "Employee 1",
				EmployeeID:     uuid.New().String(),
				EmployeeEmail:  "employee1@example.com",
				CardType:       models.CardTypeVirtual,
				Status:         models.CardStatusActive,
				Balance:        100.00,
				SpendingLimit:  floatPtr(1000.00),
				DailyLimit:     floatPtr(200.00),
				MonthlyLimit:   floatPtr(5000.00),
				ExpiryDate:     time.Now().AddDate(3, 0, 0),
				CVVHash:        "test-cvv-hash-1",
				LastFour:       "1111",
				CreatedAt:      time.Now().Add(-2 * time.Hour), // Older card
				UpdatedAt:      time.Now().Add(-2 * time.Hour),
			},
			{
				ID:             uuid.New(),
				CompanyID:      companyID,
				CardNumber:     "4222222222222222",
				CardHolderName: "Employee 2",
				EmployeeID:     uuid.New().String(),
				EmployeeEmail:  "employee2@example.com",
				CardType:       models.CardTypeVirtual,
				Status:         models.CardStatusActive,
				Balance:        200.00,
				SpendingLimit:  floatPtr(2000.00),
				DailyLimit:     floatPtr(300.00),
				MonthlyLimit:   floatPtr(6000.00),
				ExpiryDate:     time.Now().AddDate(3, 0, 0),
				CVVHash:        "test-cvv-hash-2",
				LastFour:       "2222",
				CreatedAt:      time.Now().Add(-1 * time.Hour), // Newer card
				UpdatedAt:      time.Now().Add(-1 * time.Hour),
			},
		}

		err = clientRepo.CreateCardsInBatch(ctx, cards)
		require.NoError(t, err)

		retrievedCards, err := cardRepo.GetCardsByCompanyID(ctx, companyID)
		require.NoError(t, err)
		require.Len(t, retrievedCards, 2)

		assert.Equal(t, "employee2@example.com", retrievedCards[0].EmployeeEmail)
		assert.Equal(t, "employee1@example.com", retrievedCards[1].EmployeeEmail)

		expectedCards := make(map[string]*models.Card)
		for _, card := range cards {
			expectedCards[card.ID.String()] = card
		}

		for _, retrievedCard := range retrievedCards {
			expectedCard, exists := expectedCards[retrievedCard.ID.String()]
			require.True(t, exists, "Retrieved card ID not found in expected cards")

			assert.Equal(t, expectedCard.CompanyID, retrievedCard.CompanyID)
			assert.Equal(t, expectedCard.CardNumber, retrievedCard.CardNumber)
			assert.Equal(t, expectedCard.CardHolderName, retrievedCard.CardHolderName)
			assert.Equal(t, expectedCard.EmployeeID, retrievedCard.EmployeeID)
			assert.Equal(t, expectedCard.EmployeeEmail, retrievedCard.EmployeeEmail)
			assert.Equal(t, expectedCard.CardType, retrievedCard.CardType)
			assert.Equal(t, expectedCard.Status, retrievedCard.Status)
			assert.Equal(t, expectedCard.Balance, retrievedCard.Balance)
			assert.Equal(t, expectedCard.SpendingLimit, retrievedCard.SpendingLimit)
			assert.Equal(t, expectedCard.DailyLimit, retrievedCard.DailyLimit)
			assert.Equal(t, expectedCard.MonthlyLimit, retrievedCard.MonthlyLimit)
			assert.Equal(t, expectedCard.LastFour, retrievedCard.LastFour)
			assert.WithinDuration(t, expectedCard.CreatedAt, retrievedCard.CreatedAt, time.Second)
			assert.WithinDuration(t, expectedCard.UpdatedAt, retrievedCard.UpdatedAt, time.Second)
		}
	})

	t.Run("company_with_no_cards", func(t *testing.T) {
		companyID := uuid.New()

		retrievedCards, err := cardRepo.GetCardsByCompanyID(ctx, companyID)
		require.NoError(t, err)
		assert.Empty(t, retrievedCards, "Should return empty slice for company with no cards")
	})
}

func floatPtr(v float64) *float64 {
	return &v
}
