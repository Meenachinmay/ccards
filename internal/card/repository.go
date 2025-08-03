package card

import (
	"context"
	"database/sql"

	"github.com/google/uuid"

	"ccards/pkg/models"
)

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{
		db: db,
	}
}

func (r *repository) GetCardsByCompanyID(ctx context.Context, companyID uuid.UUID) ([]*models.Card, error) {
	query := `
		SELECT id, company_id, card_number, card_holder_name, employee_id, employee_email, 
		       card_type, status, balance, spending_limit, daily_limit, monthly_limit, 
		       expiry_date, cvv_hash, last_four, created_at, updated_at, blocked_at, blocked_reason
		FROM cards
		WHERE company_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []*models.Card
	for rows.Next() {
		var card models.Card
		err := rows.Scan(
			&card.ID, &card.CompanyID, &card.CardNumber, &card.CardHolderName,
			&card.EmployeeID, &card.EmployeeEmail, &card.CardType, &card.Status,
			&card.Balance, &card.SpendingLimit, &card.DailyLimit, &card.MonthlyLimit,
			&card.ExpiryDate, &card.CVVHash, &card.LastFour, &card.CreatedAt,
			&card.UpdatedAt, &card.BlockedAt, &card.BlockedReason,
		)
		if err != nil {
			return nil, err
		}
		cards = append(cards, &card)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return cards, nil
}

func (r *repository) UpdateSpendingLimit(ctx context.Context, id uuid.UUID, spendingLimit int) (*models.Card, error) {
	query := `
		UPDATE cards
		SET spending_limit = $2, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING id, company_id, card_number, card_holder_name, employee_id, employee_email, 
		       card_type, status, balance, spending_limit, daily_limit, monthly_limit, 
		       expiry_date, cvv_hash, last_four, created_at, updated_at, blocked_at, blocked_reason
	`

	var card models.Card
	err := r.db.QueryRowContext(ctx, query, id, spendingLimit).Scan(
		&card.ID, &card.CompanyID, &card.CardNumber, &card.CardHolderName,
		&card.EmployeeID, &card.EmployeeEmail, &card.CardType, &card.Status,
		&card.Balance, &card.SpendingLimit, &card.DailyLimit, &card.MonthlyLimit,
		&card.ExpiryDate, &card.CVVHash, &card.LastFour, &card.CreatedAt,
		&card.UpdatedAt, &card.BlockedAt, &card.BlockedReason,
	)
	if err != nil {
		return nil, err
	}

	return &card, nil
}

func (r *repository) GetCardByCompanyIDAndCardID(ctx context.Context, companyID uuid.UUID, cardID uuid.UUID) (*models.Card, error) {
	query := `
		SELECT id, company_id, card_number, card_holder_name, employee_id, employee_email, 
		       card_type, status, balance, spending_limit, daily_limit, monthly_limit, 
		       expiry_date, cvv_hash, last_four, created_at, updated_at, blocked_at, blocked_reason
		FROM cards
		WHERE company_id = $1 AND id = $2
		ORDER BY created_at DESC
	`

	var card models.Card
	err := r.db.QueryRowContext(ctx, query, companyID, cardID).Scan(
		&card.ID, &card.CompanyID, &card.CardNumber, &card.CardHolderName,
		&card.EmployeeID, &card.EmployeeEmail, &card.CardType, &card.Status,
		&card.Balance, &card.SpendingLimit, &card.DailyLimit, &card.MonthlyLimit,
		&card.ExpiryDate, &card.CVVHash, &card.LastFour, &card.CreatedAt,
		&card.UpdatedAt, &card.BlockedAt, &card.BlockedReason,
	)
	if err != nil {
		return nil, err
	}

	return &card, nil
}
