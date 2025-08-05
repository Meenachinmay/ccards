package repository

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ccards/internal/card"
	"ccards/internal/client"
	"ccards/internal/transaction"
	"ccards/pkg/models"
	"ccards/tests/setup"
)

func setupTestCompanyAndCard(t *testing.T, ctx context.Context, clientRepo client.Repository) (*models.Company, *models.Card) {
	companyID := uuid.New()
	company := &models.Company{
		ID:       companyID,
		ClientID: uuid.New(),
		Name:     "Test Transaction Company",
		Email:    "test-tx-company-" + uuid.New().String() + "@example.com",
		Password: "hashed_password",
		Address:  "123 Test St",
		Phone:    "123-456-7890",
		Status:   models.CompanyStatusActive,
	}
	err := clientRepo.CreateCompany(ctx, company)
	require.NoError(t, err)

	cardID := uuid.New()
	uniqueCardNumber := "41111111" + uuid.New().String()[0:8]
	card := &models.Card{
		ID:             cardID,
		CompanyID:      companyID,
		CardNumber:     uniqueCardNumber,
		CardHolderName: "Test Employee",
		EmployeeID:     uuid.New().String(),
		EmployeeEmail:  "test-employee-" + uuid.New().String() + "@example.com",
		CardType:       models.CardTypeVirtual,
		Status:         models.CardStatusActive,
		Balance:        1000.00,
		SpendingLimit:  createFloatPtr(5000.00),
		DailyLimit:     createFloatPtr(1000.00),
		MonthlyLimit:   createFloatPtr(10000.00),
		ExpiryDate:     time.Now().AddDate(3, 0, 0),
		CVVHash:        "test-cvv-hash",
		LastFour:       uniqueCardNumber[len(uniqueCardNumber)-4:],
	}

	cards := []*models.Card{card}
	err = clientRepo.CreateCardsInBatch(ctx, cards)
	require.NoError(t, err)

	return company, card
}

func TestCreateTransaction(t *testing.T) {
	helper := setup.NewTestHelper(t)
	txRepo := transaction.NewRepository(helper.DB)
	clientRepo := client.NewRepository(helper.DB)
	ctx := context.Background()

	t.Run("create_transaction_success", func(t *testing.T) {
		company, card := setupTestCompanyAndCard(t, ctx, clientRepo)

		tx, err := txRepo.BeginTx(ctx)
		require.NoError(t, err)
		defer tx.Rollback()

		merchantName := "Test Merchant"
		merchantCategory := "Restaurant"
		txn := &models.Transaction{
			ID:               uuid.New(),
			CardID:           card.ID,
			CompanyID:        company.ID,
			TransactionType:  models.TransactionTypePurchase,
			Amount:           100.50,
			MerchantName:     &merchantName,
			MerchantCategory: &merchantCategory,
			Description:      "Test transaction",
			Status:           models.TransactionStatusPending,
		}

		err = txRepo.CreateTransaction(ctx, tx, txn)
		require.NoError(t, err)

		assert.NotZero(t, txn.CreatedAt)
		assert.NotZero(t, txn.UpdatedAt)
		assert.Equal(t, models.TransactionStatusPending, txn.Status)

		err = tx.Commit()
		require.NoError(t, err)

		// Verify transaction was created
		retrievedTxn, err := txRepo.GetTransactionByID(ctx, txn.ID)
		require.NoError(t, err)
		assert.Equal(t, txn.ID, retrievedTxn.ID)
		assert.Equal(t, txn.Amount, retrievedTxn.Amount)
	})

	t.Run("create_transaction_with_nil_merchant_info", func(t *testing.T) {
		company, card := setupTestCompanyAndCard(t, ctx, clientRepo)

		tx, err := txRepo.BeginTx(ctx)
		require.NoError(t, err)
		defer tx.Rollback()

		txn := &models.Transaction{
			ID:               uuid.New(),
			CardID:           card.ID,
			CompanyID:        company.ID,
			TransactionType:  models.TransactionTypeCharge,
			Amount:           50.00,
			MerchantName:     nil,
			MerchantCategory: nil,
			Description:      "Card charge",
			Status:           models.TransactionStatusCompleted,
		}

		err = txRepo.CreateTransaction(ctx, tx, txn)
		require.NoError(t, err)

		err = tx.Commit()
		require.NoError(t, err)
	})
}

func TestGetTransactionByID(t *testing.T) {
	helper := setup.NewTestHelper(t)
	txRepo := transaction.NewRepository(helper.DB)
	clientRepo := client.NewRepository(helper.DB)
	ctx := context.Background()

	t.Run("transaction_exists", func(t *testing.T) {
		company, card := setupTestCompanyAndCard(t, ctx, clientRepo)

		// Create transaction
		tx, err := txRepo.BeginTx(ctx)
		require.NoError(t, err)

		merchantName := "Test Merchant"
		merchantCategory := "Retail"
		txn := &models.Transaction{
			ID:               uuid.New(),
			CardID:           card.ID,
			CompanyID:        company.ID,
			TransactionType:  models.TransactionTypePurchase,
			Amount:           75.25,
			MerchantName:     &merchantName,
			MerchantCategory: &merchantCategory,
			Description:      "Test purchase",
			Status:           models.TransactionStatusCompleted,
		}

		err = txRepo.CreateTransaction(ctx, tx, txn)
		require.NoError(t, err)

		err = tx.Commit()
		require.NoError(t, err)

		// Retrieve transaction
		retrievedTxn, err := txRepo.GetTransactionByID(ctx, txn.ID)
		require.NoError(t, err)
		require.NotNil(t, retrievedTxn)

		assert.Equal(t, txn.ID, retrievedTxn.ID)
		assert.Equal(t, txn.CardID, retrievedTxn.CardID)
		assert.Equal(t, txn.CompanyID, retrievedTxn.CompanyID)
		assert.Equal(t, txn.TransactionType, retrievedTxn.TransactionType)
		assert.Equal(t, txn.Amount, retrievedTxn.Amount)
		assert.Equal(t, *txn.MerchantName, *retrievedTxn.MerchantName)
		assert.Equal(t, *txn.MerchantCategory, *retrievedTxn.MerchantCategory)
		assert.Equal(t, txn.Description, retrievedTxn.Description)
		assert.Equal(t, txn.Status, retrievedTxn.Status)
	})

	t.Run("transaction_not_exists", func(t *testing.T) {
		txID := uuid.New()
		retrievedTxn, err := txRepo.GetTransactionByID(ctx, txID)
		require.Error(t, err)
		assert.Nil(t, retrievedTxn)
	})
}

func TestGetTransactionsByCardID(t *testing.T) {
	helper := setup.NewTestHelper(t)
	txRepo := transaction.NewRepository(helper.DB)
	clientRepo := client.NewRepository(helper.DB)
	ctx := context.Background()

	t.Run("card_with_multiple_transactions", func(t *testing.T) {
		company, card := setupTestCompanyAndCard(t, ctx, clientRepo)

		// Create multiple transactions
		tx, err := txRepo.BeginTx(ctx)
		require.NoError(t, err)

		transactions := []models.Transaction{
			{
				ID:              uuid.New(),
				CardID:          card.ID,
				CompanyID:       company.ID,
				TransactionType: models.TransactionTypePurchase,
				Amount:          100.00,
				Description:     "Transaction 1",
				Status:          models.TransactionStatusCompleted,
			},
			{
				ID:              uuid.New(),
				CardID:          card.ID,
				CompanyID:       company.ID,
				TransactionType: models.TransactionTypePurchase,
				Amount:          200.00,
				Description:     "Transaction 2",
				Status:          models.TransactionStatusCompleted,
			},
			{
				ID:              uuid.New(),
				CardID:          card.ID,
				CompanyID:       company.ID,
				TransactionType: models.TransactionTypeCharge,
				Amount:          50.00,
				Description:     "Transaction 3",
				Status:          models.TransactionStatusCompleted,
			},
		}

		for i := range transactions {
			err = txRepo.CreateTransaction(ctx, tx, &transactions[i])
			require.NoError(t, err)
			time.Sleep(10 * time.Millisecond)
		}

		err = tx.Commit()
		require.NoError(t, err)

		// Test pagination
		retrievedTxns, err := txRepo.GetTransactionsByCardID(ctx, card.ID, 2, 0)
		require.NoError(t, err)
		assert.Len(t, retrievedTxns, 2)

		// Should be ordered by created_at DESC
		assert.Equal(t, transactions[2].ID, retrievedTxns[0].ID)
		assert.Equal(t, transactions[1].ID, retrievedTxns[1].ID)

		// Test offset
		retrievedTxns, err = txRepo.GetTransactionsByCardID(ctx, card.ID, 2, 2)
		require.NoError(t, err)
		assert.Len(t, retrievedTxns, 1)
		assert.Equal(t, transactions[0].ID, retrievedTxns[0].ID)
	})

	t.Run("card_with_no_transactions", func(t *testing.T) {
		_, card := setupTestCompanyAndCard(t, ctx, clientRepo)

		retrievedTxns, err := txRepo.GetTransactionsByCardID(ctx, card.ID, 10, 0)
		require.NoError(t, err)
		assert.Empty(t, retrievedTxns)
	})
}

func TestGetTransactionsByCompanyID(t *testing.T) {
	helper := setup.NewTestHelper(t)
	txRepo := transaction.NewRepository(helper.DB)
	clientRepo := client.NewRepository(helper.DB)
	_ = card.NewRepository(helper.DB)
	ctx := context.Background()

	t.Run("company_with_multiple_cards_and_transactions", func(t *testing.T) {
		company, card1 := setupTestCompanyAndCard(t, ctx, clientRepo)

		uniqueCardNumber2 := "42222222" + uuid.New().String()[0:8]
		card2 := &models.Card{
			ID:             uuid.New(),
			CompanyID:      company.ID,
			CardNumber:     uniqueCardNumber2,
			CardHolderName: "Employee 2",
			EmployeeID:     uuid.New().String(),
			EmployeeEmail:  "employee2-" + uuid.New().String() + "@example.com",
			CardType:       models.CardTypeVirtual,
			Status:         models.CardStatusActive,
			Balance:        500.00,
			SpendingLimit:  createFloatPtr(2000.00),
			DailyLimit:     createFloatPtr(500.00),
			MonthlyLimit:   createFloatPtr(5000.00),
			ExpiryDate:     time.Now().AddDate(3, 0, 0),
			CVVHash:        "test-cvv-hash-2",
			LastFour:       uniqueCardNumber2[len(uniqueCardNumber2)-4:],
		}
		err := clientRepo.CreateCardsInBatch(ctx, []*models.Card{card2})
		require.NoError(t, err)

		tx, err := txRepo.BeginTx(ctx)
		require.NoError(t, err)

		transactions := []models.Transaction{
			{
				ID:              uuid.New(),
				CardID:          card1.ID,
				CompanyID:       company.ID,
				TransactionType: models.TransactionTypePurchase,
				Amount:          100.00,
				Description:     "Card 1 Transaction 1",
				Status:          models.TransactionStatusCompleted,
			},
			{
				ID:              uuid.New(),
				CardID:          card2.ID,
				CompanyID:       company.ID,
				TransactionType: models.TransactionTypePurchase,
				Amount:          200.00,
				Description:     "Card 2 Transaction 1",
				Status:          models.TransactionStatusCompleted,
			},
			{
				ID:              uuid.New(),
				CardID:          card1.ID,
				CompanyID:       company.ID,
				TransactionType: models.TransactionTypePurchase,
				Amount:          150.00,
				Description:     "Card 1 Transaction 2",
				Status:          models.TransactionStatusCompleted,
			},
		}

		for i := range transactions {
			err = txRepo.CreateTransaction(ctx, tx, &transactions[i])
			require.NoError(t, err)
			time.Sleep(10 * time.Millisecond)
		}

		err = tx.Commit()
		require.NoError(t, err)

		retrievedTxns, err := txRepo.GetTransactionsByCompanyID(ctx, company.ID, 10, 0)
		require.NoError(t, err)
		assert.Len(t, retrievedTxns, 3)

		assert.Equal(t, transactions[2].ID, retrievedTxns[0].ID)
		assert.Equal(t, transactions[1].ID, retrievedTxns[1].ID)
		assert.Equal(t, transactions[0].ID, retrievedTxns[2].ID)
	})
}

func TestUpdateTransactionStatus(t *testing.T) {
	helper := setup.NewTestHelper(t)
	txRepo := transaction.NewRepository(helper.DB)
	clientRepo := client.NewRepository(helper.DB)
	ctx := context.Background()

	t.Run("update_status_success", func(t *testing.T) {
		company, card := setupTestCompanyAndCard(t, ctx, clientRepo)

		tx, err := txRepo.BeginTx(ctx)
		require.NoError(t, err)

		txn := &models.Transaction{
			ID:              uuid.New(),
			CardID:          card.ID,
			CompanyID:       company.ID,
			TransactionType: models.TransactionTypePurchase,
			Amount:          100.00,
			Description:     "Test transaction",
			Status:          models.TransactionStatusPending,
		}

		err = txRepo.CreateTransaction(ctx, tx, txn)
		require.NoError(t, err)

		err = txRepo.UpdateTransactionStatus(ctx, tx, txn.ID, models.TransactionStatusCompleted)
		require.NoError(t, err)

		err = tx.Commit()
		require.NoError(t, err)

		retrievedTxn, err := txRepo.GetTransactionByID(ctx, txn.ID)
		require.NoError(t, err)
		assert.Equal(t, models.TransactionStatusCompleted, retrievedTxn.Status)
		assert.NotNil(t, retrievedTxn.ProcessedAt)
	})

	t.Run("update_non_existent_transaction", func(t *testing.T) {
		tx, err := txRepo.BeginTx(ctx)
		require.NoError(t, err)
		defer tx.Rollback()

		err = txRepo.UpdateTransactionStatus(ctx, tx, uuid.New(), models.TransactionStatusCompleted)
		require.NoError(t, err)
	})
}

func TestGetTotalSpentToday(t *testing.T) {
	helper := setup.NewTestHelper(t)
	txRepo := transaction.NewRepository(helper.DB)
	clientRepo := client.NewRepository(helper.DB)
	ctx := context.Background()

	t.Run("calculate_today_spending", func(t *testing.T) {
		company, card := setupTestCompanyAndCard(t, ctx, clientRepo)

		tx, err := txRepo.BeginTx(ctx)
		require.NoError(t, err)

		todayTransactions := []models.Transaction{
			{
				ID:              uuid.New(),
				CardID:          card.ID,
				CompanyID:       company.ID,
				TransactionType: models.TransactionTypePurchase,
				Amount:          100.00,
				Description:     "Today Transaction 1",
				Status:          models.TransactionStatusCompleted,
			},
			{
				ID:              uuid.New(),
				CardID:          card.ID,
				CompanyID:       company.ID,
				TransactionType: models.TransactionTypePurchase,
				Amount:          200.00,
				Description:     "Today Transaction 2",
				Status:          models.TransactionStatusCompleted,
			},
			{
				ID:              uuid.New(),
				CardID:          card.ID,
				CompanyID:       company.ID,
				TransactionType: models.TransactionTypePurchase,
				Amount:          50.00,
				Description:     "Today Transaction 3",
				Status:          models.TransactionStatusPending,
			},
			{
				ID:              uuid.New(),
				CardID:          card.ID,
				CompanyID:       company.ID,
				TransactionType: models.TransactionTypeCharge,
				Amount:          75.00,
				Description:     "Today Charge",
				Status:          models.TransactionStatusCompleted,
			},
		}

		for i := range todayTransactions {
			err = txRepo.CreateTransaction(ctx, tx, &todayTransactions[i])
			require.NoError(t, err)
		}

		err = tx.Commit()
		require.NoError(t, err)

		total, err := txRepo.GetTotalSpentToday(ctx, card.ID)
		require.NoError(t, err)
		assert.Equal(t, 300.00, total)
	})

	t.Run("no_transactions_today", func(t *testing.T) {
		_, card := setupTestCompanyAndCard(t, ctx, clientRepo)

		total, err := txRepo.GetTotalSpentToday(ctx, card.ID)
		require.NoError(t, err)
		assert.Equal(t, 0.0, total)
	})
}

func TestGetTotalSpentThisMonth(t *testing.T) {
	helper := setup.NewTestHelper(t)
	txRepo := transaction.NewRepository(helper.DB)
	clientRepo := client.NewRepository(helper.DB)
	ctx := context.Background()

	t.Run("calculate_monthly_spending", func(t *testing.T) {
		company, card := setupTestCompanyAndCard(t, ctx, clientRepo)

		tx, err := txRepo.BeginTx(ctx)
		require.NoError(t, err)

		monthTransactions := []models.Transaction{
			{
				ID:              uuid.New(),
				CardID:          card.ID,
				CompanyID:       company.ID,
				TransactionType: models.TransactionTypePurchase,
				Amount:          500.00,
				Description:     "Month Transaction 1",
				Status:          models.TransactionStatusCompleted,
			},
			{
				ID:              uuid.New(),
				CardID:          card.ID,
				CompanyID:       company.ID,
				TransactionType: models.TransactionTypePurchase,
				Amount:          300.00,
				Description:     "Month Transaction 2",
				Status:          models.TransactionStatusCompleted,
			},
			{
				ID:              uuid.New(),
				CardID:          card.ID,
				CompanyID:       company.ID,
				TransactionType: models.TransactionTypePurchase,
				Amount:          100.00,
				Description:     "Month Transaction 3",
				Status:          models.TransactionStatusFailed,
			},
		}

		for i := range monthTransactions {
			err = txRepo.CreateTransaction(ctx, tx, &monthTransactions[i])
			require.NoError(t, err)
		}

		err = tx.Commit()
		require.NoError(t, err)

		total, err := txRepo.GetTotalSpentThisMonth(ctx, card.ID)
		require.NoError(t, err)
		assert.Equal(t, 800.00, total) // 500 + 300
	})
}

func TestUpdateCardBalance(t *testing.T) {
	helper := setup.NewTestHelper(t)
	txRepo := transaction.NewRepository(helper.DB)
	clientRepo := client.NewRepository(helper.DB)
	_ = card.NewRepository(helper.DB)
	ctx := context.Background()

	t.Run("update_balance_success", func(t *testing.T) {
		_, card := setupTestCompanyAndCard(t, ctx, clientRepo)

		tx, err := txRepo.BeginTx(ctx)
		require.NoError(t, err)

		// Initial balance is 1000.00
		err = txRepo.UpdateCardBalance(ctx, tx, card.ID, 100.50)
		require.NoError(t, err)

		err = tx.Commit()
		require.NoError(t, err)

		// Verify balance
		balance, err := txRepo.GetCardBalance(ctx, card.ID)
		require.NoError(t, err)
		assert.Equal(t, 899.50, balance)
	})

	t.Run("insufficient_balance", func(t *testing.T) {
		_, card := setupTestCompanyAndCard(t, ctx, clientRepo)

		tx, err := txRepo.BeginTx(ctx)
		require.NoError(t, err)
		defer tx.Rollback()

		// Try to deduct more than available balance
		err = txRepo.UpdateCardBalance(ctx, tx, card.ID, 2000.00)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient balance")
	})

	t.Run("concurrent_balance_updates", func(t *testing.T) {
		_, card := setupTestCompanyAndCard(t, ctx, clientRepo)

		var wg sync.WaitGroup
		wg.Add(2)

		errors := make(chan error, 2)

		go func() {
			defer wg.Done()
			tx, err := txRepo.BeginTx(ctx)
			if err != nil {
				errors <- err
				return
			}
			defer tx.Rollback()

			err = txRepo.UpdateCardBalance(ctx, tx, card.ID, 100.00)
			if err != nil {
				errors <- err
				return
			}

			err = tx.Commit()
			if err != nil {
				errors <- err
				return
			}
			errors <- nil
		}()

		go func() {
			defer wg.Done()
			tx, err := txRepo.BeginTx(ctx)
			if err != nil {
				errors <- err
				return
			}
			defer tx.Rollback()

			err = txRepo.UpdateCardBalance(ctx, tx, card.ID, 100.00)
			if err != nil {
				errors <- err
				return
			}

			err = tx.Commit()
			if err != nil {
				errors <- err
				return
			}
			errors <- nil
		}()

		wg.Wait()
		close(errors)

		successCount := 0
		for err := range errors {
			if err == nil {
				successCount++
			}
		}
		assert.Equal(t, 2, successCount)

		balance, err := txRepo.GetCardBalance(ctx, card.ID)
		require.NoError(t, err)
		assert.Equal(t, 800.00, balance)
	})
}

func TestGetCardBalance(t *testing.T) {
	helper := setup.NewTestHelper(t)
	txRepo := transaction.NewRepository(helper.DB)
	clientRepo := client.NewRepository(helper.DB)
	ctx := context.Background()

	t.Run("get_balance_success", func(t *testing.T) {
		_, card := setupTestCompanyAndCard(t, ctx, clientRepo)

		balance, err := txRepo.GetCardBalance(ctx, card.ID)
		require.NoError(t, err)
		assert.Equal(t, 1000.00, balance)
	})

	t.Run("card_not_exists", func(t *testing.T) {
		balance, err := txRepo.GetCardBalance(ctx, uuid.New())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "card not found")
		assert.Equal(t, 0.0, balance)
	})
}

func TestBeginTx(t *testing.T) {
	helper := setup.NewTestHelper(t)
	txRepo := transaction.NewRepository(helper.DB)
	ctx := context.Background()

	t.Run("begin_transaction_success", func(t *testing.T) {
		tx, err := txRepo.BeginTx(ctx)
		require.NoError(t, err)
		require.NotNil(t, tx)

		// Should be able to rollback
		err = tx.Rollback()
		require.NoError(t, err)
	})

	t.Run("multiple_transactions", func(t *testing.T) {
		tx1, err := txRepo.BeginTx(ctx)
		require.NoError(t, err)
		defer tx1.Rollback()

		tx2, err := txRepo.BeginTx(ctx)
		require.NoError(t, err)
		defer tx2.Rollback()

		// Both transactions should be independent
		assert.NotEqual(t, tx1, tx2)
	})
}

// Helper function to create a pointer to a float64
func createFloatPtr(v float64) *float64 {
	return &v
}
