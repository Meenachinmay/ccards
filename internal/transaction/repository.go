package transaction

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"ccards/pkg/models"
	"github.com/google/uuid"
)

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) CreateTransaction(ctx context.Context, tx *sql.Tx, transaction *models.Transaction) error {
	now := time.Now()
	transaction.CreatedAt = now
	transaction.UpdatedAt = now

	query := `
        INSERT INTO transactions (
            id, card_id, company_id, transaction_type, amount,
            merchant_name, merchant_category, description, status, created_at, updated_at
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
        RETURNING created_at, updated_at`

	err := tx.QueryRowContext(
		ctx,
		query,
		transaction.ID,
		transaction.CardID,
		transaction.CompanyID,
		transaction.TransactionType,
		transaction.Amount,
		transaction.MerchantName,
		transaction.MerchantCategory,
		transaction.Description,
		transaction.Status,
		transaction.CreatedAt,
		transaction.UpdatedAt,
	).Scan(&transaction.CreatedAt, &transaction.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}

	return nil
}

func (r *repository) GetTransactionByID(ctx context.Context, id uuid.UUID) (*models.Transaction, error) {
	var transaction models.Transaction
	query := `
        SELECT id, card_id, company_id, transaction_type, amount,
               merchant_name, merchant_category, description, status,
               processed_at, created_at, updated_at
        FROM transactions
        WHERE id = $1`

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&transaction.ID,
		&transaction.CardID,
		&transaction.CompanyID,
		&transaction.TransactionType,
		&transaction.Amount,
		&transaction.MerchantName,
		&transaction.MerchantCategory,
		&transaction.Description,
		&transaction.Status,
		&transaction.ProcessedAt,
		&transaction.CreatedAt,
		&transaction.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("transaction not found")
		}
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	return &transaction, nil
}

func (r *repository) GetTransactionsByCardID(ctx context.Context, cardID uuid.UUID, limit, offset int) ([]*models.Transaction, error) {
	query := `
        SELECT id, card_id, company_id, transaction_type, amount,
               merchant_name, merchant_category, description, status,
               processed_at, created_at, updated_at
        FROM transactions
        WHERE card_id = $1
        ORDER BY created_at DESC
        LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, query, cardID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get transactions: %w", err)
	}
	defer rows.Close()

	var transactions []*models.Transaction
	for rows.Next() {
		var transaction models.Transaction
		err := rows.Scan(
			&transaction.ID,
			&transaction.CardID,
			&transaction.CompanyID,
			&transaction.TransactionType,
			&transaction.Amount,
			&transaction.MerchantName,
			&transaction.MerchantCategory,
			&transaction.Description,
			&transaction.Status,
			&transaction.ProcessedAt,
			&transaction.CreatedAt,
			&transaction.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan transaction: %w", err)
		}
		transactions = append(transactions, &transaction)
	}

	return transactions, nil
}

func (r *repository) GetTransactionsByCompanyID(ctx context.Context, companyID uuid.UUID, limit, offset int) ([]*models.Transaction, error) {
	query := `
        SELECT id, card_id, company_id, transaction_type, amount,
               merchant_name, merchant_category, description, status,
               processed_at, created_at, updated_at
        FROM transactions
        WHERE company_id = $1
        ORDER BY created_at DESC
        LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, query, companyID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get transactions: %w", err)
	}
	defer rows.Close()

	var transactions []*models.Transaction
	for rows.Next() {
		var transaction models.Transaction
		err := rows.Scan(
			&transaction.ID,
			&transaction.CardID,
			&transaction.CompanyID,
			&transaction.TransactionType,
			&transaction.Amount,
			&transaction.MerchantName,
			&transaction.MerchantCategory,
			&transaction.Description,
			&transaction.Status,
			&transaction.ProcessedAt,
			&transaction.CreatedAt,
			&transaction.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan transaction: %w", err)
		}
		transactions = append(transactions, &transaction)
	}

	return transactions, nil
}

func (r *repository) UpdateTransactionStatus(ctx context.Context, tx *sql.Tx, id uuid.UUID, status string) error {
	query := `
        UPDATE transactions
        SET status = $2, processed_at = $3
        WHERE id = $1`

	processedAt := time.Now()
	_, err := tx.ExecContext(ctx, query, id, status, processedAt)
	if err != nil {
		return fmt.Errorf("failed to update transaction status: %w", err)
	}

	return nil
}

func (r *repository) GetTotalSpentToday(ctx context.Context, cardID uuid.UUID) (float64, error) {
	var total sql.NullFloat64
	query := `
        SELECT COALESCE(SUM(amount), 0)
        FROM transactions
        WHERE card_id = $1
        AND transaction_type = $2
        AND status = $3
        AND DATE(created_at) = CURRENT_DATE`

	err := r.db.QueryRowContext(
		ctx,
		query,
		cardID,
		models.TransactionTypePurchase,
		models.TransactionStatusCompleted,
	).Scan(&total)

	if err != nil {
		return 0, fmt.Errorf("failed to get daily total: %w", err)
	}

	return total.Float64, nil
}

func (r *repository) GetTotalSpentThisMonth(ctx context.Context, cardID uuid.UUID) (float64, error) {
	var total sql.NullFloat64
	query := `
        SELECT COALESCE(SUM(amount), 0)
        FROM transactions
        WHERE card_id = $1
        AND transaction_type = $2
        AND status = $3
        AND DATE_TRUNC('month', created_at) = DATE_TRUNC('month', CURRENT_DATE)`

	err := r.db.QueryRowContext(
		ctx,
		query,
		cardID,
		models.TransactionTypePurchase,
		models.TransactionStatusCompleted,
	).Scan(&total)

	if err != nil {
		return 0, fmt.Errorf("failed to get monthly total: %w", err)
	}

	return total.Float64, nil
}

func (r *repository) UpdateCardBalance(ctx context.Context, tx *sql.Tx, cardID uuid.UUID, amount float64) error {
	var currentBalance float64
	lockQuery := `SELECT balance FROM cards WHERE id = $1 FOR UPDATE`

	err := tx.QueryRowContext(ctx, lockQuery, cardID).Scan(&currentBalance)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("card not found")
		}
		return fmt.Errorf("failed to lock card for update: %w", err)
	}

	if currentBalance < amount {
		return fmt.Errorf("insufficient balance")
	}

	updateQuery := `UPDATE cards SET balance = balance - $2 WHERE id = $1`
	_, err = tx.ExecContext(ctx, updateQuery, cardID, amount)
	if err != nil {
		return fmt.Errorf("failed to update card balance: %w", err)
	}

	return nil
}

func (r *repository) GetCardBalance(ctx context.Context, cardID uuid.UUID) (float64, error) {
	var balance float64
	query := `SELECT balance FROM cards WHERE id = $1`

	err := r.db.QueryRowContext(ctx, query, cardID).Scan(&balance)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("card not found")
		}
		return 0, fmt.Errorf("failed to get card balance: %w", err)
	}

	return balance, nil
}

func (r *repository) BeginTx(ctx context.Context) (*sql.Tx, error) {
	return r.db.BeginTx(ctx, nil)
}
