package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"ccards/internal/card"
	"ccards/pkg/models"
)

type MockCardRepository struct {
	mock.Mock
}

func (m *MockCardRepository) GetCardsByCompanyID(ctx context.Context, companyID uuid.UUID) ([]*models.Card, error) {
	args := m.Called(ctx, companyID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Card), args.Error(1)
}

func TestGetCardsByCompanyID(t *testing.T) {
	mockRepo := new(MockCardRepository)
	svc := card.NewService(mockRepo)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		companyID := uuid.New()
		expectedCards := []*models.Card{
			{
				ID:             uuid.New(),
				CompanyID:      companyID,
				CardNumber:     "4111111111111111",
				CardHolderName: "Employee 1",
				EmployeeEmail:  "employee1@example.com",
				CardType:       models.CardTypeVirtual,
				Status:         models.CardStatusActive,
			},
			{
				ID:             uuid.New(),
				CompanyID:      companyID,
				CardNumber:     "4222222222222222",
				CardHolderName: "Employee 2",
				EmployeeEmail:  "employee2@example.com",
				CardType:       models.CardTypeVirtual,
				Status:         models.CardStatusActive,
			},
		}

		mockRepo.On("GetCardsByCompanyID", ctx, companyID).Return(expectedCards, nil).Once()

		cards, err := svc.GetCardsByCompanyID(ctx, companyID)
		require.NoError(t, err)
		require.NotNil(t, cards)
		require.Len(t, cards, 2)

		assert.Equal(t, expectedCards, cards)
		mockRepo.AssertExpectations(t)
	})

	t.Run("empty_result", func(t *testing.T) {
		companyID := uuid.New()
		mockRepo.On("GetCardsByCompanyID", ctx, companyID).Return([]*models.Card{}, nil).Once()

		cards, err := svc.GetCardsByCompanyID(ctx, companyID)
		require.NoError(t, err)
		assert.Empty(t, cards)
		mockRepo.AssertExpectations(t)
	})

	t.Run("repository_error", func(t *testing.T) {
		companyID := uuid.New()
		expectedErr := assert.AnError
		mockRepo.On("GetCardsByCompanyID", ctx, companyID).Return(nil, expectedErr).Once()

		cards, err := svc.GetCardsByCompanyID(ctx, companyID)
		require.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Nil(t, cards)
		mockRepo.AssertExpectations(t)
	})
}