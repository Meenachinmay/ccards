package models

import (
	"github.com/google/uuid"
	"time"
)

const (
	CompanyStatusActive    = "active"
	CompanyStatusSuspended = "suspended"
	CompanyStatusInactive  = "inactive"
)

const (
	CardToIssueStatusPending   = "pending"
	CardToIssueStatusGenerated = "generated"
)

const (
	CardTypeVirtual  = "virtual"
	CardTypePhysical = "physical"

	CardStatusActive    = "active"
	CardStatusBlocked   = "blocked"
	CardStatusExpired   = "expired"
	CardStatusCancelled = "cancelled"
)

const (
	TransactionTypePurchase = "purchase"
	TransactionTypeCharge   = "charge"

	TransactionStatusPending   = "pending"
	TransactionStatusCompleted = "completed"
	TransactionStatusFailed    = "failed"
)

type Company struct {
	ID        uuid.UUID `json:"id"`
	ClientID  uuid.UUID `json:"client_id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	Address   string    `json:"address,omitempty"`
	Phone     string    `json:"phone,omitempty"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CardToIssue struct {
	ID            uuid.UUID `json:"id" db:"id"`
	ClientID      uuid.UUID `json:"client_id" db:"client_id"`
	CardID        uuid.UUID `json:"card_id" db:"card_id"`
	EmployeeID    uuid.UUID `json:"employee_id" db:"employee_id"`
	EmployeeEmail string    `json:"employee_email" db:"employee_email"`
	Status        string    `json:"status" db:"status"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

type Card struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	CompanyID      uuid.UUID  `json:"company_id" db:"company_id"`
	CardNumber     string     `json:"card_number" db:"card_number"`
	CardHolderName string     `json:"card_holder_name" db:"card_holder_name"`
	EmployeeID     string     `json:"employee_id" db:"employee_id"`
	EmployeeEmail  string     `json:"employee_email" db:"employee_email"`
	CardType       string     `json:"card_type" db:"card_type"`
	Status         string     `json:"status" db:"status"`
	Balance        float64    `json:"balance" db:"balance"`
	SpendingLimit  *float64   `json:"spending_limit" db:"spending_limit"`
	DailyLimit     *float64   `json:"daily_limit" db:"daily_limit"`
	MonthlyLimit   *float64   `json:"monthly_limit" db:"monthly_limit"`
	ExpiryDate     time.Time  `json:"expiry_date" db:"expiry_date"`
	CVVHash        string     `json:"cvv_hash" db:"cvv_hash"`
	LastFour       string     `json:"last_four" db:"last_four"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
	BlockedAt      *time.Time `json:"blocked_at" db:"blocked_at"`
	BlockedReason  *string    `json:"blocked_reason" db:"blocked_reason"`
}

type Transaction struct {
	ID               uuid.UUID  `json:"id" db:"id"`
	CardID           uuid.UUID  `json:"card_id" db:"card_id"`
	CompanyID        uuid.UUID  `json:"company_id" db:"company_id"`
	TransactionType  string     `json:"transaction_type" db:"transaction_type"`
	Amount           float64    `json:"amount" db:"amount"`
	MerchantName     *string    `json:"merchant_name" db:"merchant_name"`
	MerchantCategory *string    `json:"merchant_category" db:"merchant_category"`
	Description      string     `json:"description" db:"description"`
	Status           string     `json:"status" db:"status"`
	ProcessedAt      *time.Time `json:"processed_at" db:"processed_at"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at" db:"updated_at"`
}

type SpendingControl struct {
	ID           uuid.UUID   `json:"id" db:"id"`
	CardID       uuid.UUID   `json:"card_id" db:"card_id"`
	ControlType  string      `json:"control_type" db:"control_type"`
	ControlValue interface{} `json:"control_value" db:"control_value"`
	IsActive     bool        `json:"is_active" db:"is_active"`
	CreatedAt    time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at" db:"updated_at"`
}
