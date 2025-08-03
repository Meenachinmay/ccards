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
