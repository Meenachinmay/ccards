package response

import (
	"time"

	"github.com/google/uuid"
)

type Company struct {
	ID        uuid.UUID `json:"id"`
	ClientID  uuid.UUID `json:"client_id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Address   string    `json:"address,omitempty"`
	Phone     string    `json:"phone,omitempty"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type RegisterCompanyResponse struct {
	Company     Company `json:"company"`
	Credentials struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	} `json:"credentials"`
}

type LoginResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	TokenType    string    `json:"token_type"`
	Company      Company   `json:"company"`
}

type RefreshTokenResponse struct {
	AccessToken string    `json:"access_token"`
	ExpiresAt   time.Time `json:"expires_at"`
	TokenType   string    `json:"token_type"`
}
