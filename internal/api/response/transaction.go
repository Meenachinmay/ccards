package response

import (
	"github.com/google/uuid"
	"time"
)

type Transaction struct {
	ID               uuid.UUID  `json:"id"`
	CardID           uuid.UUID  `json:"card_id"`
	CompanyID        uuid.UUID  `json:"company_id"`
	TransactionType  string     `json:"transaction_type"`
	Amount           float64    `json:"amount"`
	MerchantName     *string    `json:"merchant_name,omitempty"`
	MerchantCategory *string    `json:"merchant_category,omitempty"`
	Description      string     `json:"description"`
	Status           string     `json:"status"`
	ProcessedAt      *time.Time `json:"processed_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

type TransactionResponse struct {
	Transaction      Transaction `json:"transaction"`
	RemainingBalance float64     `json:"remaining_balance"`
	CardLastFour     string      `json:"card_last_four"`
}

type TransactionListResponse struct {
	Transactions []Transaction `json:"transactions"`
	Total        int           `json:"total"`
	Page         int           `json:"page"`
	PageSize     int           `json:"page_size"`
}
