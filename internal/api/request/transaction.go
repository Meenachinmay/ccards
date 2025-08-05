package request

import "github.com/google/uuid"

type Transaction struct {
	CompanyID        uuid.UUID `json:"company_id" binding:"required"`
	CardID           uuid.UUID `json:"card_id" binding:"required"`
	Amount           float64   `json:"amount" binding:"required"`
	MerchantCategory string    `json:"merchant_category" binding:"required"`
}
