package client

import (
	"context"

	"github.com/google/uuid"

	"ccards/internal/api/request"
	"ccards/internal/api/response"
	"ccards/pkg/models"
)

type Repository interface {
	CreateCompany(ctx context.Context, company *models.Company) error
	GetCompanyByID(ctx context.Context, id uuid.UUID) (*models.Company, error)
	GetCompanyByEmail(ctx context.Context, email string) (*models.Company, error)

	CreateCardsToIssue(ctx context.Context, cards []*models.CardToIssue) error
	GetCardsToIssueByClientID(ctx context.Context, clientID uuid.UUID) ([]*models.CardToIssue, error)
	UpdateCardToIssueStatus(ctx context.Context, id uuid.UUID, status string) error

	GetPendingCardsToIssue(ctx context.Context, companyID uuid.UUID) ([]*models.CardToIssue, error)
	CreateCardsInBatch(ctx context.Context, cards []*models.Card) error
	UpdateCardsToIssueStatusBatch(ctx context.Context, ids []uuid.UUID, status string) error
}

type Service interface {
	RegisterCompany(ctx context.Context, req *request.RegisterCompany) (*response.RegisterCompanyResponse, error)
	Login(ctx context.Context, req *request.LoginCompany) (*response.LoginResponse, error)
	RefreshToken(ctx context.Context, refreshToken string) (*response.RefreshTokenResponse, error)
	GetCompanyByID(ctx context.Context, id uuid.UUID) (*models.Company, error)
	GetCompanyByEmail(ctx context.Context, email string) (*models.Company, error)
	ProcessCardCSVUpload(ctx context.Context, clientID uuid.UUID, csvData []byte) error
	GetCardsToIssueByClientID(ctx context.Context, clientID uuid.UUID) ([]*models.CardToIssue, error)

	IssueNewCards(ctx context.Context, companyID uuid.UUID) (int, error)
}
