package card

import (
	"context"

	"github.com/google/uuid"

	"ccards/pkg/models"
)

type Repository interface {
	GetCardsByCompanyID(ctx context.Context, companyID uuid.UUID) ([]*models.Card, error)
}

type Service interface {
	GetCardsByCompanyID(ctx context.Context, companyID uuid.UUID) ([]*models.Card, error)
}
