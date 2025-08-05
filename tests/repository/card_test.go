package repository

import (
	"context"
	"encoding/json"
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

func TestGetCardByCompanyIDAndCardID(t *testing.T) {
	helper := setup.NewTestHelper(t)
	cardRepo := card.NewRepository(helper.DB)
	clientRepo := client.NewRepository(helper.DB)
	ctx := context.Background()

	t.Run("card_exists", func(t *testing.T) {
		companyID := uuid.New()
		company := &models.Company{
			ID:       companyID,
			ClientID: uuid.New(),
			Name:     "Test Company",
			Email:    "test-get-card-company@example.com",
			Password: "hashed_password",
			Address:  "123 Test St",
			Phone:    "123-456-7890",
			Status:   models.CompanyStatusActive,
		}
		err := clientRepo.CreateCompany(ctx, company)
		require.NoError(t, err)

		cardID := uuid.New()
		card := &models.Card{
			ID:             cardID,
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
		}

		cards := []*models.Card{card}
		err = clientRepo.CreateCardsInBatch(ctx, cards)
		require.NoError(t, err)

		retrievedCard, err := cardRepo.GetCardByCompanyIDAndCardID(ctx, companyID, cardID)
		require.NoError(t, err)
		require.NotNil(t, retrievedCard)

		assert.Equal(t, cardID, retrievedCard.ID)
		assert.Equal(t, companyID, retrievedCard.CompanyID)
		assert.Equal(t, card.CardNumber, retrievedCard.CardNumber)
		assert.Equal(t, card.CardHolderName, retrievedCard.CardHolderName)
		assert.Equal(t, card.EmployeeEmail, retrievedCard.EmployeeEmail)
		assert.Equal(t, card.CardType, retrievedCard.CardType)
		assert.Equal(t, card.Status, retrievedCard.Status)
		assert.Equal(t, card.Balance, retrievedCard.Balance)
		assert.Equal(t, *card.SpendingLimit, *retrievedCard.SpendingLimit)
	})

	t.Run("card_not_exists", func(t *testing.T) {
		companyID := uuid.New()
		cardID := uuid.New()

		retrievedCard, err := cardRepo.GetCardByCompanyIDAndCardID(ctx, companyID, cardID)
		require.Error(t, err)
		assert.Nil(t, retrievedCard)
	})
}

func TestUpdateSpendingLimit(t *testing.T) {
	helper := setup.NewTestHelper(t)
	cardRepo := card.NewRepository(helper.DB)
	clientRepo := client.NewRepository(helper.DB)
	ctx := context.Background()

	t.Run("update_success", func(t *testing.T) {
		companyID := uuid.New()
		company := &models.Company{
			ID:       companyID,
			ClientID: uuid.New(),
			Name:     "Test Company",
			Email:    "test-update-limit-company@example.com",
			Password: "hashed_password",
			Address:  "123 Test St",
			Phone:    "123-456-7890",
			Status:   models.CompanyStatusActive,
		}
		err := clientRepo.CreateCompany(ctx, company)
		require.NoError(t, err)

		cardID := uuid.New()
		initialSpendingLimit := 1000.00
		card := &models.Card{
			ID:             cardID,
			CompanyID:      companyID,
			CardNumber:     "4111111111111111",
			CardHolderName: "Employee 1",
			EmployeeID:     uuid.New().String(),
			EmployeeEmail:  "employee1@example.com",
			CardType:       models.CardTypeVirtual,
			Status:         models.CardStatusActive,
			Balance:        100.00,
			SpendingLimit:  floatPtr(initialSpendingLimit),
			DailyLimit:     floatPtr(200.00),
			MonthlyLimit:   floatPtr(5000.00),
			ExpiryDate:     time.Now().AddDate(3, 0, 0),
			CVVHash:        "test-cvv-hash-1",
			LastFour:       "1111",
		}

		cards := []*models.Card{card}
		err = clientRepo.CreateCardsInBatch(ctx, cards)
		require.NoError(t, err)

		newSpendingLimit := 2000
		updatedCard, err := cardRepo.UpdateSpendingLimit(ctx, cardID, newSpendingLimit)
		require.NoError(t, err)
		require.NotNil(t, updatedCard)

		assert.Equal(t, cardID, updatedCard.ID)
		assert.Equal(t, companyID, updatedCard.CompanyID)
		assert.Equal(t, float64(newSpendingLimit), *updatedCard.SpendingLimit)

		retrievedCard, err := cardRepo.GetCardByCompanyIDAndCardID(ctx, companyID, cardID)
		require.NoError(t, err)
		require.NotNil(t, retrievedCard)
		assert.Equal(t, float64(newSpendingLimit), *retrievedCard.SpendingLimit)
	})

	t.Run("card_not_exists", func(t *testing.T) {
		cardID := uuid.New()
		newSpendingLimit := 2000

		updatedCard, err := cardRepo.UpdateSpendingLimit(ctx, cardID, newSpendingLimit)
		require.Error(t, err)
		assert.Nil(t, updatedCard)
	})
}

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

func TestUpdateSpendingControl(t *testing.T) {
	helper := setup.NewTestHelper(t)
	cardRepo := card.NewRepository(helper.DB)
	clientRepo := client.NewRepository(helper.DB)
	ctx := context.Background()

	t.Run("create_new_control_success", func(t *testing.T) {
		companyID := uuid.New()
		company := &models.Company{
			ID:       companyID,
			ClientID: uuid.New(),
			Name:     "Test Company",
			Email:    "test-spending-control-company@example.com",
			Password: "hashed_password",
			Address:  "123 Test St",
			Phone:    "123-456-7890",
			Status:   models.CompanyStatusActive,
		}
		err := clientRepo.CreateCompany(ctx, company)
		require.NoError(t, err)

		cardID := uuid.New()
		card := &models.Card{
			ID:             cardID,
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
		}

		cards := []*models.Card{card}
		err = clientRepo.CreateCardsInBatch(ctx, cards)
		require.NoError(t, err)

		controlType := "merchant_category"
		controlValue := map[string]interface{}{
			"allowed_categories": []string{"food"},
			"blocked_categories": []string{},
		}

		err = cardRepo.UpdateSpendingControl(ctx, cardID, controlType, controlValue)
		require.NoError(t, err)

		// Verify the control was created by querying the database directly
		query := `
			SELECT control_type, control_value, is_active
			FROM spending_controls
			WHERE card_id = $1 AND control_type = $2
		`
		var retrievedType string
		var controlValueJSON []byte
		var isActive bool

		err = helper.DB.QueryRowContext(ctx, query, cardID, controlType).Scan(&retrievedType, &controlValueJSON, &isActive)
		require.NoError(t, err)

		assert.Equal(t, controlType, retrievedType)
		assert.True(t, isActive)

		// Verify the control value
		var retrievedValue map[string]interface{}
		err = json.Unmarshal(controlValueJSON, &retrievedValue)
		require.NoError(t, err)

		allowedCategories, ok := retrievedValue["allowed_categories"].([]interface{})
		require.True(t, ok)
		require.Len(t, allowedCategories, 1)
		assert.Equal(t, "food", allowedCategories[0])

		blockedCategories, ok := retrievedValue["blocked_categories"].([]interface{})
		require.True(t, ok)
		require.Len(t, blockedCategories, 0)
	})

	t.Run("update_existing_control_success", func(t *testing.T) {
		companyID := uuid.New()
		company := &models.Company{
			ID:       companyID,
			ClientID: uuid.New(),
			Name:     "Test Company",
			Email:    "test-update-control-company@example.com",
			Password: "hashed_password",
			Address:  "123 Test St",
			Phone:    "123-456-7890",
			Status:   models.CompanyStatusActive,
		}
		err := clientRepo.CreateCompany(ctx, company)
		require.NoError(t, err)

		cardID := uuid.New()
		card := &models.Card{
			ID:             cardID,
			CompanyID:      companyID,
			CardNumber:     "4333333333333333",
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
			LastFour:       "3333",
		}

		cards := []*models.Card{card}
		err = clientRepo.CreateCardsInBatch(ctx, cards)
		require.NoError(t, err)

		// First, create an initial control
		controlType := "merchant_category"
		initialControlValue := map[string]interface{}{
			"allowed_categories": []string{"food"},
			"blocked_categories": []string{},
		}

		err = cardRepo.UpdateSpendingControl(ctx, cardID, controlType, initialControlValue)
		require.NoError(t, err)

		// Now update the control
		updatedControlValue := map[string]interface{}{
			"allowed_categories": []string{"food", "grocery"},
			"blocked_categories": []string{"entertainment"},
		}

		err = cardRepo.UpdateSpendingControl(ctx, cardID, controlType, updatedControlValue)
		require.NoError(t, err)

		// Verify the control was updated
		query := `
			SELECT control_type, control_value, is_active
			FROM spending_controls
			WHERE card_id = $1 AND control_type = $2
		`
		var retrievedType string
		var controlValueJSON []byte
		var isActive bool

		err = helper.DB.QueryRowContext(ctx, query, cardID, controlType).Scan(&retrievedType, &controlValueJSON, &isActive)
		require.NoError(t, err)

		assert.Equal(t, controlType, retrievedType)
		assert.True(t, isActive)

		// Verify the updated control value
		var retrievedValue map[string]interface{}
		err = json.Unmarshal(controlValueJSON, &retrievedValue)
		require.NoError(t, err)

		allowedCategories, ok := retrievedValue["allowed_categories"].([]interface{})
		require.True(t, ok)
		require.Len(t, allowedCategories, 2)
		assert.Contains(t, allowedCategories, "food")
		assert.Contains(t, allowedCategories, "grocery")

		blockedCategories, ok := retrievedValue["blocked_categories"].([]interface{})
		require.True(t, ok)
		require.Len(t, blockedCategories, 1)
		assert.Equal(t, "entertainment", blockedCategories[0])
	})

	t.Run("card_not_exists", func(t *testing.T) {
		cardID := uuid.New()
		controlType := "merchant_category"
		controlValue := map[string]interface{}{
			"allowed_categories": []string{"food"},
			"blocked_categories": []string{},
		}

		err := cardRepo.UpdateSpendingControl(ctx, cardID, controlType, controlValue)
		require.Error(t, err)
	})
}

func floatPtr(v float64) *float64 {
	return &v
}
