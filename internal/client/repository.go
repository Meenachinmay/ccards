package client

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"ccards/pkg/models"
)

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) CreateCompany(ctx context.Context, company *models.Company) error {
	query := `
		INSERT INTO companies (id, client_id, name, email, password, address, phone, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING created_at, updated_at`

	err := r.db.QueryRowContext(ctx, query,
		company.ID, company.ClientID, company.Name, company.Email, company.Password, company.Address, company.Phone, company.Status,
	).Scan(&company.CreatedAt, &company.UpdatedAt)

	if err != nil {
		return err
	}

	return nil
}

func (r *repository) GetCompanyByID(ctx context.Context, id uuid.UUID) (*models.Company, error) {
	company := &models.Company{}
	query := `
		SELECT id, client_id, name, email, address, phone, status, created_at, updated_at
		FROM companies
		WHERE id = $1`

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&company.ID, &company.ClientID, &company.Name, &company.Email, &company.Address, &company.Phone, &company.Status,
		&company.CreatedAt, &company.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return company, nil
}

func (r *repository) GetCompanyByEmail(ctx context.Context, email string) (*models.Company, error) {
	company := &models.Company{}
	query := `
		SELECT id, client_id, name, email, password, address, phone, status, created_at, updated_at
		FROM companies
		WHERE email = $1`

	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&company.ID, &company.ClientID, &company.Name, &company.Email, &company.Password, &company.Address, &company.Phone, &company.Status,
		&company.CreatedAt, &company.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return company, nil
}

func (r *repository) CreateCardsToIssue(ctx context.Context, cards []*models.CardToIssue) error {
	if len(cards) == 0 {
		return nil
	}

	query := `
        INSERT INTO cards_to_issue (
            id, client_id, card_id, employee_id, employee_email, status, created_at, updated_at
        ) VALUES `

	values := make([]string, 0, len(cards))
	args := make([]interface{}, 0, len(cards)*8)

	for i, card := range cards {
		values = append(values, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)",
			i*8+1, i*8+2, i*8+3, i*8+4, i*8+5, i*8+6, i*8+7, i*8+8))

		args = append(args,
			card.ID,
			card.ClientID,
			card.CardID,
			card.EmployeeID,
			card.EmployeeEmail,
			card.Status,
			card.CreatedAt,
			card.UpdatedAt,
		)
	}

	query += strings.Join(values, ", ")

	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to create cards to issue: %w", err)
	}

	return nil
}

func (r *repository) GetCardsToIssueByClientID(ctx context.Context, clientID uuid.UUID) ([]*models.CardToIssue, error) {
	query := `
        SELECT id, client_id, card_id, employee_id, employee_email, status, created_at, updated_at
        FROM cards_to_issue
        WHERE client_id = $1
        ORDER BY created_at DESC
    `

	rows, err := r.db.QueryContext(ctx, query, clientID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cards to issue: %w", err)
	}
	defer rows.Close()

	var cards []*models.CardToIssue
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
		if err != nil {
			return nil, fmt.Errorf("failed to scan card to issue: %w", err)
		}
		cards = append(cards, card)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}

	return cards, nil
}

func (r *repository) UpdateCardToIssueStatus(ctx context.Context, id uuid.UUID, status string) error {
	query := `
        UPDATE cards_to_issue
        SET status = $2, updated_at = CURRENT_TIMESTAMP
        WHERE id = $1
    `

	result, err := r.db.ExecContext(ctx, query, id, status)
	if err != nil {
		return fmt.Errorf("failed to update card to issue status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("card to issue not found")
	}

	return nil
}
