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
