package repository

import (
	"context"
	"fmt"
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

		savedCards, err := repo.GetCardsToIssueByClientID(ctx, clientID)
		require.NoError(t, err)
		require.Len(t, savedCards, 2)

		expectedCards := make(map[string]*models.CardToIssue)
		for _, card := range cards {
			expectedCards[card.ID.String()] = card
		}

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

		retrievedCards, err := repo.GetCardsToIssueByClientID(ctx, clientID)
		require.NoError(t, err)
		require.Len(t, retrievedCards, 2)

		assert.Equal(t, models.CardToIssueStatusGenerated, retrievedCards[0].Status)
		assert.Equal(t, models.CardToIssueStatusPending, retrievedCards[1].Status)

		expectedCards := make(map[string]*models.CardToIssue)
		for _, card := range cards {
			expectedCards[card.ID.String()] = card
		}

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

		err = repo.UpdateCardToIssueStatus(ctx, cardID, models.CardToIssueStatusGenerated)
		require.NoError(t, err)

		cards, err := repo.GetCardsToIssueByClientID(ctx, clientID)
		require.NoError(t, err)
		require.Len(t, cards, 1)

		updatedCard := cards[0]
		assert.Equal(t, cardID, updatedCard.ID)
		assert.Equal(t, models.CardToIssueStatusGenerated, updatedCard.Status)

		assert.True(t, updatedCard.UpdatedAt.After(card.UpdatedAt) ||
			updatedCard.UpdatedAt.Equal(card.UpdatedAt),
			"UpdatedAt should be equal to or after the original timestamp")
	})

	t.Run("update_non_existing_card", func(t *testing.T) {
		nonExistingID := uuid.New()

		err := repo.UpdateCardToIssueStatus(ctx, nonExistingID, models.CardToIssueStatusGenerated)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "card to issue not found")
	})

	t.Run("update_with_invalid_status", func(t *testing.T) {
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

		invalidStatus := "invalid_status"
		err = repo.UpdateCardToIssueStatus(ctx, cardID, invalidStatus)

		require.Error(t, err)
	})
}

func TestGetPendingCardsToIssue(t *testing.T) {
	helper := setup.NewTestHelper(t)
	repo := client.NewRepository(helper.DB)
	ctx := context.Background()

	t.Run("company_with_pending_cards", func(t *testing.T) {
		companyID := uuid.New()

		pendingCards := []*models.CardToIssue{
			{
				ID:            uuid.New(),
				ClientID:      companyID,
				CardID:        uuid.New(),
				EmployeeID:    uuid.New(),
				EmployeeEmail: "pending1@example.com",
				Status:        models.CardToIssueStatusPending,
				CreatedAt:     time.Now().Add(-2 * time.Hour), // Older card
				UpdatedAt:     time.Now().Add(-2 * time.Hour),
			},
			{
				ID:            uuid.New(),
				ClientID:      companyID,
				CardID:        uuid.New(),
				EmployeeID:    uuid.New(),
				EmployeeEmail: "pending2@example.com",
				Status:        models.CardToIssueStatusPending,
				CreatedAt:     time.Now().Add(-1 * time.Hour), // Newer card
				UpdatedAt:     time.Now().Add(-1 * time.Hour),
			},
		}

		generatedCard := &models.CardToIssue{
			ID:            uuid.New(),
			ClientID:      companyID,
			CardID:        uuid.New(),
			EmployeeID:    uuid.New(),
			EmployeeEmail: "generated@example.com",
			Status:        models.CardToIssueStatusGenerated,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}

		allCards := append(pendingCards, generatedCard)
		err := repo.CreateCardsToIssue(ctx, allCards)
		require.NoError(t, err)

		rows, err := helper.DB.QueryContext(ctx, `
			SELECT id, client_id, card_id, employee_id, employee_email, status, created_at, updated_at
			FROM cards_to_issue
			WHERE client_id = $1 AND status = $2
			ORDER BY created_at ASC
		`, companyID, models.CardToIssueStatusPending)
		require.NoError(t, err)
		defer rows.Close()

		var retrievedCards []*models.CardToIssue
		for rows.Next() {
			card := &models.CardToIssue{}
			err := rows.Scan(
				&card.ID,
				&card.ClientID,
				&card.CardID,
				&card.EmployeeID,
				&card.EmployeeEmail,
				&card.Status,
				&card.CreatedAt,
				&card.UpdatedAt,
			)
			require.NoError(t, err)
			retrievedCards = append(retrievedCards, card)
		}
		require.NoError(t, rows.Err())

		require.Len(t, retrievedCards, 2, "Should only return pending cards")

		assert.Equal(t, "pending1@example.com", retrievedCards[0].EmployeeEmail)
		assert.Equal(t, "pending2@example.com", retrievedCards[1].EmployeeEmail)

		for _, card := range retrievedCards {
			assert.Equal(t, models.CardToIssueStatusPending, card.Status)
		}
	})

	t.Run("company_with_no_pending_cards", func(t *testing.T) {
		companyID := uuid.New()

		generatedCard := &models.CardToIssue{
			ID:            uuid.New(),
			ClientID:      companyID,
			CardID:        uuid.New(),
			EmployeeID:    uuid.New(),
			EmployeeEmail: "only-generated@example.com",
			Status:        models.CardToIssueStatusGenerated,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}

		err := repo.CreateCardsToIssue(ctx, []*models.CardToIssue{generatedCard})
		require.NoError(t, err)

		rows, err := helper.DB.QueryContext(ctx, `
			SELECT COUNT(*)
			FROM cards_to_issue
			WHERE client_id = $1 AND status = $2
		`, companyID, models.CardToIssueStatusPending)
		require.NoError(t, err)
		defer rows.Close()

		var count int
		require.True(t, rows.Next())
		err = rows.Scan(&count)
		require.NoError(t, err)
		require.NoError(t, rows.Err())

		assert.Equal(t, 0, count, "Should return 0 for company with no pending cards")
	})

	t.Run("company_with_no_cards", func(t *testing.T) {
		companyID := uuid.New()

		rows, err := helper.DB.QueryContext(ctx, `
			SELECT COUNT(*)
			FROM cards_to_issue
			WHERE client_id = $1 AND status = $2
		`, companyID, models.CardToIssueStatusPending)
		require.NoError(t, err)
		defer rows.Close()

		var count int
		require.True(t, rows.Next())
		err = rows.Scan(&count)
		require.NoError(t, err)
		require.NoError(t, rows.Err())

		assert.Equal(t, 0, count, "Should return 0 for company with no cards")
	})
}

func TestCreateCardsInBatch(t *testing.T) {
	helper := setup.NewTestHelper(t)
	repo := client.NewRepository(helper.DB)
	ctx := context.Background()

	t.Run("success_single_card", func(t *testing.T) {
		companyID := uuid.New()
		cardID := uuid.New()

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
		err := repo.CreateCompany(ctx, company)
		require.NoError(t, err)

		// Verify company was created
		var companyExists bool
		err = helper.DB.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM companies WHERE id = $1)", companyID).Scan(&companyExists)
		require.NoError(t, err)
		require.True(t, companyExists, "Company should exist in the database")

		card := &models.Card{
			ID:             cardID,
			CompanyID:      companyID,
			CardNumber:     "4111111111111111",
			CardHolderName: "Test Employee",
			EmployeeID:     uuid.New().String(),
			EmployeeEmail:  "employee1@example.com",
			CardType:       models.CardTypeVirtual,
			Status:         models.CardStatusActive,
			Balance:        0.00,
			ExpiryDate:     time.Now().AddDate(3, 0, 0),
			CVVHash:        "test-cvv-hash",
			LastFour:       "1111",
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}

		err = repo.CreateCardsInBatch(ctx, []*models.Card{card})
		require.NoError(t, err)

		var count int
		err = helper.DB.QueryRowContext(ctx, "SELECT COUNT(*) FROM cards WHERE id = $1", cardID).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 1, count, "Card should be created in the database")

		var retrievedCard models.Card
		err = helper.DB.QueryRowContext(ctx, `
			SELECT id, company_id, card_number, card_holder_name, employee_id, employee_email, 
			       card_type, status, balance, last_four
			FROM cards WHERE id = $1
		`, cardID).Scan(
			&retrievedCard.ID,
			&retrievedCard.CompanyID,
			&retrievedCard.CardNumber,
			&retrievedCard.CardHolderName,
			&retrievedCard.EmployeeID,
			&retrievedCard.EmployeeEmail,
			&retrievedCard.CardType,
			&retrievedCard.Status,
			&retrievedCard.Balance,
			&retrievedCard.LastFour,
		)
		require.NoError(t, err)

		assert.Equal(t, card.ID, retrievedCard.ID)
		assert.Equal(t, card.CompanyID, retrievedCard.CompanyID)
		assert.Equal(t, card.CardNumber, retrievedCard.CardNumber)
		assert.Equal(t, card.CardHolderName, retrievedCard.CardHolderName)
		assert.Equal(t, card.EmployeeID, retrievedCard.EmployeeID)
		assert.Equal(t, card.EmployeeEmail, retrievedCard.EmployeeEmail)
		assert.Equal(t, card.CardType, retrievedCard.CardType)
		assert.Equal(t, card.Status, retrievedCard.Status)
		assert.Equal(t, card.Balance, retrievedCard.Balance)
		assert.Equal(t, card.LastFour, retrievedCard.LastFour)
	})

	t.Run("success_multiple_cards", func(t *testing.T) {
		companyID := uuid.New()

		company := &models.Company{
			ID:       companyID,
			ClientID: uuid.New(),
			Name:     "Test Company Multiple Cards",
			Email:    "test-multiple-cards@example.com",
			Password: "hashed_password",
			Address:  "456 Test St",
			Phone:    "123-456-7890",
			Status:   models.CompanyStatusActive,
		}
		err := repo.CreateCompany(ctx, company)
		require.NoError(t, err)

		var companyExists bool
		err = helper.DB.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM companies WHERE id = $1)", companyID).Scan(&companyExists)
		require.NoError(t, err)
		require.True(t, companyExists, "Company should exist in the database")

		cards := []*models.Card{
			{
				ID:             uuid.New(),
				CompanyID:      companyID,
				CardNumber:     "4222222222222222",
				CardHolderName: "Employee 2",
				EmployeeID:     uuid.New().String(),
				EmployeeEmail:  "employee2@example.com",
				CardType:       models.CardTypeVirtual,
				Status:         models.CardStatusActive,
				Balance:        0.00,
				ExpiryDate:     time.Now().AddDate(3, 0, 0),
				CVVHash:        "test-cvv-hash-2",
				LastFour:       "2222",
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
			},
			{
				ID:             uuid.New(),
				CompanyID:      companyID,
				CardNumber:     "4333333333333333",
				CardHolderName: "Employee 3",
				EmployeeID:     uuid.New().String(),
				EmployeeEmail:  "employee3@example.com",
				CardType:       models.CardTypeVirtual,
				Status:         models.CardStatusActive,
				Balance:        0.00,
				ExpiryDate:     time.Now().AddDate(3, 0, 0),
				CVVHash:        "test-cvv-hash-3",
				LastFour:       "3333",
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
			},
		}

		err = repo.CreateCardsInBatch(ctx, cards)
		require.NoError(t, err)

		var count int
		err = helper.DB.QueryRowContext(ctx, "SELECT COUNT(*) FROM cards WHERE company_id = $1", companyID).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 2, count, "Both cards should be created in the database")

		for _, card := range cards {
			var retrievedCard models.Card
			err = helper.DB.QueryRowContext(ctx, `
				SELECT id, company_id, card_number, card_holder_name, employee_id, employee_email, 
					   card_type, status, balance, last_four
				FROM cards WHERE id = $1
			`, card.ID).Scan(
				&retrievedCard.ID,
				&retrievedCard.CompanyID,
				&retrievedCard.CardNumber,
				&retrievedCard.CardHolderName,
				&retrievedCard.EmployeeID,
				&retrievedCard.EmployeeEmail,
				&retrievedCard.CardType,
				&retrievedCard.Status,
				&retrievedCard.Balance,
				&retrievedCard.LastFour,
			)
			require.NoError(t, err)

			assert.Equal(t, card.ID, retrievedCard.ID)
			assert.Equal(t, card.CompanyID, retrievedCard.CompanyID)
			assert.Equal(t, card.CardNumber, retrievedCard.CardNumber)
			assert.Equal(t, card.CardHolderName, retrievedCard.CardHolderName)
			assert.Equal(t, card.EmployeeID, retrievedCard.EmployeeID)
			assert.Equal(t, card.EmployeeEmail, retrievedCard.EmployeeEmail)
			assert.Equal(t, card.CardType, retrievedCard.CardType)
			assert.Equal(t, card.Status, retrievedCard.Status)
			assert.Equal(t, card.Balance, retrievedCard.Balance)
			assert.Equal(t, card.LastFour, retrievedCard.LastFour)
		}
	})

	t.Run("empty_cards_slice", func(t *testing.T) {
		err := repo.CreateCardsInBatch(ctx, []*models.Card{})
		require.NoError(t, err, "Creating empty cards slice should not return an error")
	})
}

func TestUpdateCardsToIssueStatusBatch(t *testing.T) {
	helper := setup.NewTestHelper(t)
	repo := client.NewRepository(helper.DB)
	ctx := context.Background()

	t.Run("update_multiple_cards", func(t *testing.T) {
		companyID := uuid.New()
		cardIDs := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}

		var cards []*models.CardToIssue
		for i, id := range cardIDs {
			card := &models.CardToIssue{
				ID:            id,
				ClientID:      companyID,
				CardID:        uuid.New(),
				EmployeeID:    uuid.New(),
				EmployeeEmail: fmt.Sprintf("batch-update-%d@example.com", i),
				Status:        models.CardToIssueStatusPending,
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
			}
			cards = append(cards, card)
		}

		err := repo.CreateCardsToIssue(ctx, cards)
		require.NoError(t, err)

		err = repo.UpdateCardsToIssueStatusBatch(ctx, cardIDs, models.CardToIssueStatusGenerated)
		require.NoError(t, err)

		for _, id := range cardIDs {
			var status string
			err = helper.DB.QueryRowContext(ctx, "SELECT status FROM cards_to_issue WHERE id = $1", id).Scan(&status)
			require.NoError(t, err)
			assert.Equal(t, models.CardToIssueStatusGenerated, status)
		}
	})

	t.Run("update_subset_of_cards", func(t *testing.T) {
		companyID := uuid.New()
		allCardIDs := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}

		var cards []*models.CardToIssue
		for i, id := range allCardIDs {
			card := &models.CardToIssue{
				ID:            id,
				ClientID:      companyID,
				CardID:        uuid.New(),
				EmployeeID:    uuid.New(),
				EmployeeEmail: fmt.Sprintf("batch-subset-%d@example.com", i),
				Status:        models.CardToIssueStatusPending,
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
			}
			cards = append(cards, card)
		}

		err := repo.CreateCardsToIssue(ctx, cards)
		require.NoError(t, err)

		updateIDs := allCardIDs[:2]
		err = repo.UpdateCardsToIssueStatusBatch(ctx, updateIDs, models.CardToIssueStatusGenerated)
		require.NoError(t, err)

		for i, id := range allCardIDs {
			var status string
			err = helper.DB.QueryRowContext(ctx, "SELECT status FROM cards_to_issue WHERE id = $1", id).Scan(&status)
			require.NoError(t, err)

			if i < 2 {
				assert.Equal(t, models.CardToIssueStatusGenerated, status)
			} else {
				assert.Equal(t, models.CardToIssueStatusPending, status)
			}
		}
	})

	t.Run("empty_ids_slice", func(t *testing.T) {
		err := repo.UpdateCardsToIssueStatusBatch(ctx, []uuid.UUID{}, models.CardToIssueStatusGenerated)
		require.NoError(t, err, "Updating empty IDs slice should not return an error")
	})

	t.Run("update_with_invalid_status", func(t *testing.T) {
		cardID := uuid.New()
		card := &models.CardToIssue{
			ID:            cardID,
			ClientID:      uuid.New(),
			CardID:        uuid.New(),
			EmployeeID:    uuid.New(),
			EmployeeEmail: "batch-invalid@example.com",
			Status:        models.CardToIssueStatusPending,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}

		err := repo.CreateCardsToIssue(ctx, []*models.CardToIssue{card})
		require.NoError(t, err)

		invalidStatus := "invalid_status"
		err = repo.UpdateCardsToIssueStatusBatch(ctx, []uuid.UUID{cardID}, invalidStatus)

		require.Error(t, err)
	})
}
