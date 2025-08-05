package transaction

import (
	"ccards/pkg/models"
	"context"
	"database/sql"
	"github.com/google/uuid"
)

type Repository interface {
	// CreateTransaction Transaction operations
	CreateTransaction(ctx context.Context, tx *sql.Tx, transaction *models.Transaction) error
	GetTransactionByID(ctx context.Context, id uuid.UUID) (*models.Transaction, error)
	GetTransactionsByCardID(ctx context.Context, cardID uuid.UUID, limit, offset int) ([]*models.Transaction, error)
	GetTransactionsByCompanyID(ctx context.Context, companyID uuid.UUID, limit, offset int) ([]*models.Transaction, error)
	UpdateTransactionStatus(ctx context.Context, tx *sql.Tx, id uuid.UUID, status string) error
	GetTotalSpentToday(ctx context.Context, cardID uuid.UUID) (float64, error)
	GetTotalSpentThisMonth(ctx context.Context, cardID uuid.UUID) (float64, error)

	UpdateCardBalance(ctx context.Context, tx *sql.Tx, cardID uuid.UUID, amount float64) error
	GetCardBalance(ctx context.Context, cardID uuid.UUID) (float64, error)

	BeginTx(ctx context.Context) (*sql.Tx, error)
}

type Service interface {
	ProcessPayment(ctx context.Context, companyID, cardID uuid.UUID, amount float64, merchantCategory string) (*models.Transaction, float64, error)
	GetTransaction(ctx context.Context, id uuid.UUID) (*models.Transaction, error)
	GetCardTransactions(ctx context.Context, cardID uuid.UUID, limit, offset int) ([]*models.Transaction, error)
	GetCompanyTransactions(ctx context.Context, companyID uuid.UUID, limit, offset int) ([]*models.Transaction, error)
	GetDailySpending(ctx context.Context, cardID uuid.UUID) (float64, error)
	GetMonthlySpending(ctx context.Context, cardID uuid.UUID) (float64, error)
}
