package card

import (
	"context"

	"github.com/google/uuid"

	"ccards/pkg/models"
)

type Repository interface {
	GetCardsByCompanyID(ctx context.Context, companyID uuid.UUID) ([]*models.Card, error)
	UpdateSpendingLimit(ctx context.Context, id uuid.UUID, spendingLimit int) (*models.Card, error)
	GetCardByCompanyIDAndCardID(ctx context.Context, companyID uuid.UUID, cardID uuid.UUID) (*models.Card, error)
}

type Service interface {
	GetCardsByCompanyID(ctx context.Context, companyID uuid.UUID) ([]*models.Card, error)
	GetCardByCompanyIDAndCardID(ctx context.Context, companyID uuid.UUID, cardID uuid.UUID) (*models.Card, error)
	UpdateSpendingLimit(ctx context.Context, id uuid.UUID, spendingLimit int) (*models.Card, error)
}
