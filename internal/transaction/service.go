package transaction

import (
	"context"
	"fmt"
	"time"

	"ccards/pkg/models"
	"github.com/google/uuid"
)

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) ProcessPayment(ctx context.Context, companyID, cardID uuid.UUID, amount float64, merchantCategory string) (*models.Transaction, float64, error) {
	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	transaction := &models.Transaction{
		ID:               uuid.New(),
		CardID:           cardID,
		CompanyID:        companyID,
		TransactionType:  models.TransactionTypePurchase,
		Amount:           amount,
		MerchantCategory: &merchantCategory,
		Description:      "Card purchase",
		Status:           models.TransactionStatusPending,
	}

	// Insert transaction
	if err := s.repo.CreateTransaction(ctx, tx, transaction); err != nil {
		return nil, 0, fmt.Errorf("failed to create transaction: %w", err)
	}

	// Update card balance
	if err := s.repo.UpdateCardBalance(ctx, tx, cardID, amount); err != nil {
		return nil, 0, fmt.Errorf("failed to update card balance: %w", err)
	}

	// Update transaction status to completed
	if err := s.repo.UpdateTransactionStatus(ctx, tx, transaction.ID, models.TransactionStatusCompleted); err != nil {
		return nil, 0, fmt.Errorf("failed to update transaction status: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Get updated balance
	balance, err := s.repo.GetCardBalance(ctx, cardID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get updated balance: %w", err)
	}

	// Update transaction object with completed status and processed time
	transaction.Status = models.TransactionStatusCompleted
	now := time.Now()
	transaction.ProcessedAt = &now

	return transaction, balance, nil
}

func (s *service) GetTransaction(ctx context.Context, id uuid.UUID) (*models.Transaction, error) {
	return s.repo.GetTransactionByID(ctx, id)
}

func (s *service) GetCardTransactions(ctx context.Context, cardID uuid.UUID, limit, offset int) ([]*models.Transaction, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	return s.repo.GetTransactionsByCardID(ctx, cardID, limit, offset)
}

func (s *service) GetCompanyTransactions(ctx context.Context, companyID uuid.UUID, limit, offset int) ([]*models.Transaction, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	return s.repo.GetTransactionsByCompanyID(ctx, companyID, limit, offset)
}

func (s *service) GetDailySpending(ctx context.Context, cardID uuid.UUID) (float64, error) {
	return s.repo.GetTotalSpentToday(ctx, cardID)
}

func (s *service) GetMonthlySpending(ctx context.Context, cardID uuid.UUID) (float64, error) {
	return s.repo.GetTotalSpentThisMonth(ctx, cardID)
}
