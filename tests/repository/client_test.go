package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ccards/internal/client"
	"ccards/pkg/models"
	"ccards/tests/setup"
)

func TestCreateCompany(t *testing.T) {
	helper := setup.NewTestHelper(t)
	repo := client.NewRepository(helper.DB)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		company := &models.Company{
			ID:       uuid.New(),
			ClientID: uuid.New(),
			Name:     "Test Company",
			Email:    "test@example.com",
			Password: "hashed_password",
			Address:  "123 Test St",
			Phone:    "123-456-7890",
			Status:   models.CompanyStatusActive,
		}

		err := repo.CreateCompany(ctx, company)
		require.NoError(t, err)

		assert.False(t, company.CreatedAt.IsZero())
		assert.False(t, company.UpdatedAt.IsZero())

		savedCompany, err := repo.GetCompanyByID(ctx, company.ID)
		require.NoError(t, err)
		require.NotNil(t, savedCompany)

		assert.Equal(t, company.ID, savedCompany.ID)
		assert.Equal(t, company.ClientID, savedCompany.ClientID)
		assert.Equal(t, company.Name, savedCompany.Name)
		assert.Equal(t, company.Email, savedCompany.Email)
		assert.Equal(t, company.Address, savedCompany.Address)
		assert.Equal(t, company.Phone, savedCompany.Phone)
		assert.Equal(t, company.Status, savedCompany.Status)
		assert.WithinDuration(t, company.CreatedAt, savedCompany.CreatedAt, time.Second)
		assert.WithinDuration(t, company.UpdatedAt, savedCompany.UpdatedAt, time.Second)
	})

	t.Run("duplicate_email", func(t *testing.T) {
		company1 := &models.Company{
			ID:       uuid.New(),
			ClientID: uuid.New(),
			Name:     "Company 1",
			Email:    "duplicate@example.com",
			Password: "hashed_password1",
			Address:  "123 First St",
			Phone:    "111-111-1111",
			Status:   models.CompanyStatusActive,
		}

		err := repo.CreateCompany(ctx, company1)
		require.NoError(t, err)

		company2 := &models.Company{
			ID:       uuid.New(),
			ClientID: uuid.New(),
			Name:     "Company 2",
			Email:    "duplicate@example.com",
			Password: "hashed_password2",
			Address:  "456 Second St",
			Phone:    "222-222-2222",
			Status:   models.CompanyStatusActive,
		}

		err = repo.CreateCompany(ctx, company2)
		require.Error(t, err)
	})
}

func TestGetCompanyByID(t *testing.T) {
	helper := setup.NewTestHelper(t)
	repo := client.NewRepository(helper.DB)
	ctx := context.Background()

	t.Run("existing_company", func(t *testing.T) {
		companyID := uuid.New()
		company := &models.Company{
			ID:       companyID,
			ClientID: uuid.New(),
			Name:     "Test Company",
			Email:    "get-by-id@example.com",
			Password: "hashed_password",
			Address:  "123 Test St",
			Phone:    "123-456-7890",
			Status:   models.CompanyStatusActive,
		}

		err := repo.CreateCompany(ctx, company)
		require.NoError(t, err)

		foundCompany, err := repo.GetCompanyByID(ctx, companyID)
		require.NoError(t, err)
		require.NotNil(t, foundCompany)

		assert.Equal(t, company.ID, foundCompany.ID)
		assert.Equal(t, company.ClientID, foundCompany.ClientID)
		assert.Equal(t, company.Name, foundCompany.Name)
		assert.Equal(t, company.Email, foundCompany.Email)
		assert.Equal(t, company.Address, foundCompany.Address)
		assert.Equal(t, company.Phone, foundCompany.Phone)
		assert.Equal(t, company.Status, foundCompany.Status)
	})

	t.Run("non_existing_company", func(t *testing.T) {
		nonExistingID := uuid.New()

		foundCompany, err := repo.GetCompanyByID(ctx, nonExistingID)
		require.NoError(t, err) // No error, just nil result
		assert.Nil(t, foundCompany)
	})
}

func TestGetCompanyByEmail(t *testing.T) {
	helper := setup.NewTestHelper(t)
	repo := client.NewRepository(helper.DB)
	ctx := context.Background()

	t.Run("existing_company", func(t *testing.T) {
		email := "get-by-email@example.com"
		company := &models.Company{
			ID:       uuid.New(),
			ClientID: uuid.New(),
			Name:     "Email Test Company",
			Email:    email,
			Password: "hashed_password",
			Address:  "123 Email St",
			Phone:    "123-456-7890",
			Status:   models.CompanyStatusActive,
		}

		err := repo.CreateCompany(ctx, company)
		require.NoError(t, err)

		foundCompany, err := repo.GetCompanyByEmail(ctx, email)
		require.NoError(t, err)
		require.NotNil(t, foundCompany)

		assert.Equal(t, company.ID, foundCompany.ID)
		assert.Equal(t, company.ClientID, foundCompany.ClientID)
		assert.Equal(t, company.Name, foundCompany.Name)
		assert.Equal(t, company.Email, foundCompany.Email)
		assert.Equal(t, company.Password, foundCompany.Password) // Password should be returned
		assert.Equal(t, company.Address, foundCompany.Address)
		assert.Equal(t, company.Phone, foundCompany.Phone)
		assert.Equal(t, company.Status, foundCompany.Status)
	})

	t.Run("non_existing_email", func(t *testing.T) {
		nonExistingEmail := "non-existing@example.com"

		foundCompany, err := repo.GetCompanyByEmail(ctx, nonExistingEmail)
		require.NoError(t, err)
		assert.Nil(t, foundCompany)
	})
}

func TestCreateCardsToIssue(t *testing.T) {
	helper := setup.NewTestHelper(t)
	repo := client.NewRepository(helper.DB)
	ctx := context.Background()

	t.Run("success_single_card", func(t *testing.T) {
		clientID := uuid.New()
		cardID := uuid.New()
		employeeID := uuid.New()

		card := &models.CardToIssue{
			ID:            uuid.New(),
			ClientID:      clientID,
			CardID:        cardID,
			EmployeeID:    employeeID,
			EmployeeEmail: "employee1@example.com",
			Status:        models.CardToIssueStatusPending,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}

		err := repo.CreateCardsToIssue(ctx, []*models.CardToIssue{card})
		require.NoError(t, err)

		// Verify card was created by retrieving it
		cards, err := repo.GetCardsToIssueByClientID(ctx, clientID)
		require.NoError(t, err)
		require.Len(t, cards, 1)

		savedCard := cards[0]
		assert.Equal(t, card.ID, savedCard.ID)
		assert.Equal(t, card.ClientID, savedCard.ClientID)
		assert.Equal(t, card.CardID, savedCard.CardID)
		assert.Equal(t, card.EmployeeID, savedCard.EmployeeID)
		assert.Equal(t, card.EmployeeEmail, savedCard.EmployeeEmail)
		assert.Equal(t, card.Status, savedCard.Status)
		assert.WithinDuration(t, card.CreatedAt, savedCard.CreatedAt, time.Second)
		assert.WithinDuration(t, card.UpdatedAt, savedCard.UpdatedAt, time.Second)
	})

	t.Run("success_multiple_cards", func(t *testing.T) {
		clientID := uuid.New()

		cards := []*models.CardToIssue{
			{
				ID:            uuid.New(),
				ClientID:      clientID,
				CardID:        uuid.New(),
				EmployeeID:    uuid.New(),
				EmployeeEmail: "employee2@example.com",
				Status:        models.CardToIssueStatusPending,
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
			},
			{
				ID:            uuid.New(),
				ClientID:      clientID,
				CardID:        uuid.New(),
				EmployeeID:    uuid.New(),
				EmployeeEmail: "employee3@example.com",
				Status:        models.CardToIssueStatusPending,
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
			},
		}

		err := repo.CreateCardsToIssue(ctx, cards)
		require.NoError(t, err)

		// Verify cards were created by retrieving them
		savedCards, err := repo.GetCardsToIssueByClientID(ctx, clientID)
		require.NoError(t, err)
		require.Len(t, savedCards, 2)

		// Create a map of expected cards by ID for easier comparison
		expectedCards := make(map[string]*models.CardToIssue)
		for _, card := range cards {
			expectedCards[card.ID.String()] = card
		}

		// Verify each saved card matches an expected card
		for _, savedCard := range savedCards {
			expectedCard, exists := expectedCards[savedCard.ID.String()]
			require.True(t, exists, "Saved card ID not found in expected cards")

			assert.Equal(t, expectedCard.ClientID, savedCard.ClientID)
			assert.Equal(t, expectedCard.CardID, savedCard.CardID)
			assert.Equal(t, expectedCard.EmployeeID, savedCard.EmployeeID)
			assert.Equal(t, expectedCard.EmployeeEmail, savedCard.EmployeeEmail)
			assert.Equal(t, expectedCard.Status, savedCard.Status)
			assert.WithinDuration(t, expectedCard.CreatedAt, savedCard.CreatedAt, time.Second)
			assert.WithinDuration(t, expectedCard.UpdatedAt, savedCard.UpdatedAt, time.Second)
		}
	})

	t.Run("empty_cards_slice", func(t *testing.T) {
		err := repo.CreateCardsToIssue(ctx, []*models.CardToIssue{})
		require.NoError(t, err, "Creating empty cards slice should not return an error")
	})
}

func TestGetCardsToIssueByClientID(t *testing.T) {
	helper := setup.NewTestHelper(t)
	repo := client.NewRepository(helper.DB)
	ctx := context.Background()

	t.Run("client_with_cards", func(t *testing.T) {
		// Create a client with multiple cards
		clientID := uuid.New()

		cards := []*models.CardToIssue{
			{
				ID:            uuid.New(),
				ClientID:      clientID,
				CardID:        uuid.New(),
				EmployeeID:    uuid.New(),
				EmployeeEmail: "employee1@example.com",
				Status:        models.CardToIssueStatusPending,
				CreatedAt:     time.Now().Add(-2 * time.Hour), // Older card
				UpdatedAt:     time.Now().Add(-2 * time.Hour),
			},
			{
				ID:            uuid.New(),
				ClientID:      clientID,
				CardID:        uuid.New(),
				EmployeeID:    uuid.New(),
				EmployeeEmail: "employee2@example.com",
				Status:        models.CardToIssueStatusGenerated,
				CreatedAt:     time.Now().Add(-1 * time.Hour), // Newer card
				UpdatedAt:     time.Now().Add(-1 * time.Hour),
			},
		}

		err := repo.CreateCardsToIssue(ctx, cards)
		require.NoError(t, err)

		// Retrieve cards for the client
		retrievedCards, err := repo.GetCardsToIssueByClientID(ctx, clientID)
		require.NoError(t, err)
		require.Len(t, retrievedCards, 2)

		// Verify cards are returned in descending order by created_at (newest first)
		assert.Equal(t, models.CardToIssueStatusGenerated, retrievedCards[0].Status)
		assert.Equal(t, models.CardToIssueStatusPending, retrievedCards[1].Status)

		// Create a map of expected cards by ID for easier comparison
		expectedCards := make(map[string]*models.CardToIssue)
		for _, card := range cards {
			expectedCards[card.ID.String()] = card
		}

		// Verify each retrieved card matches an expected card
		for _, retrievedCard := range retrievedCards {
			expectedCard, exists := expectedCards[retrievedCard.ID.String()]
			require.True(t, exists, "Retrieved card ID not found in expected cards")

			assert.Equal(t, expectedCard.ClientID, retrievedCard.ClientID)
			assert.Equal(t, expectedCard.CardID, retrievedCard.CardID)
			assert.Equal(t, expectedCard.EmployeeID, retrievedCard.EmployeeID)
			assert.Equal(t, expectedCard.EmployeeEmail, retrievedCard.EmployeeEmail)
			assert.Equal(t, expectedCard.Status, retrievedCard.Status)
			assert.WithinDuration(t, expectedCard.CreatedAt, retrievedCard.CreatedAt, time.Second)
			assert.WithinDuration(t, expectedCard.UpdatedAt, retrievedCard.UpdatedAt, time.Second)
		}
	})

	t.Run("client_with_no_cards", func(t *testing.T) {
		// Use a random client ID that shouldn't have any cards
		clientID := uuid.New()

		retrievedCards, err := repo.GetCardsToIssueByClientID(ctx, clientID)
		require.NoError(t, err)
		assert.Empty(t, retrievedCards, "Should return empty slice for client with no cards")
	})
}

func TestUpdateCardToIssueStatus(t *testing.T) {
	helper := setup.NewTestHelper(t)
	repo := client.NewRepository(helper.DB)
	ctx := context.Background()

	t.Run("update_existing_card", func(t *testing.T) {
		// Create a card with pending status
		cardID := uuid.New()
		clientID := uuid.New()

		card := &models.CardToIssue{
			ID:            cardID,
			ClientID:      clientID,
			CardID:        uuid.New(),
			EmployeeID:    uuid.New(),
			EmployeeEmail: "update-test@example.com",
			Status:        models.CardToIssueStatusPending,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}

		err := repo.CreateCardsToIssue(ctx, []*models.CardToIssue{card})
		require.NoError(t, err)

		// Update the card status to generated
		err = repo.UpdateCardToIssueStatus(ctx, cardID, models.CardToIssueStatusGenerated)
		require.NoError(t, err)

		// Retrieve the card to verify the status was updated
		cards, err := repo.GetCardsToIssueByClientID(ctx, clientID)
		require.NoError(t, err)
		require.Len(t, cards, 1)

		updatedCard := cards[0]
		assert.Equal(t, cardID, updatedCard.ID)
		assert.Equal(t, models.CardToIssueStatusGenerated, updatedCard.Status)

		// Verify that updated_at was changed
		assert.True(t, updatedCard.UpdatedAt.After(card.UpdatedAt) || 
			updatedCard.UpdatedAt.Equal(card.UpdatedAt), 
			"UpdatedAt should be equal to or after the original timestamp")
	})

	t.Run("update_non_existing_card", func(t *testing.T) {
		// Try to update a card that doesn't exist
		nonExistingID := uuid.New()

		err := repo.UpdateCardToIssueStatus(ctx, nonExistingID, models.CardToIssueStatusGenerated)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "card to issue not found")
	})

	t.Run("update_with_invalid_status", func(t *testing.T) {
		// Create a card with pending status
		cardID := uuid.New()

		card := &models.CardToIssue{
			ID:            cardID,
			ClientID:      uuid.New(),
			CardID:        uuid.New(),
			EmployeeID:    uuid.New(),
			EmployeeEmail: "invalid-status@example.com",
			Status:        models.CardToIssueStatusPending,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}

		err := repo.CreateCardsToIssue(ctx, []*models.CardToIssue{card})
		require.NoError(t, err)

		// Try to update with an invalid status
		invalidStatus := "invalid_status"
		err = repo.UpdateCardToIssueStatus(ctx, cardID, invalidStatus)

		// This should fail due to the CHECK constraint in the database
		require.Error(t, err)
	})
}
