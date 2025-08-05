package card

import (
	"context"

	"github.com/google/uuid"

	"ccards/pkg/models"
)

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{
		repo: repo,
	}
}

func (s *service) GetCardsByCompanyID(ctx context.Context, companyID uuid.UUID) ([]*models.Card, error) {
	return s.repo.GetCardsByCompanyID(ctx, companyID)
}

func (s *service) GetCardByCompanyIDAndCardID(ctx context.Context, companyID uuid.UUID, cardID uuid.UUID) (*models.Card, error) {
	return s.repo.GetCardByCompanyIDAndCardID(ctx, companyID, cardID)
}

func (s *service) UpdateSpendingLimit(ctx context.Context, id uuid.UUID, spendingLimit int) (*models.Card, error) {
	return s.repo.UpdateSpendingLimit(ctx, id, spendingLimit)
}

func (s *service) UpdateSpendingControl(ctx context.Context, cardID uuid.UUID, controlType string, controlValue interface{}) error {
	return s.repo.UpdateSpendingControl(ctx, cardID, controlType, controlValue)
}
