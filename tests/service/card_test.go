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

func (m *MockCardRepository) GetCardByCompanyIDAndCardID(ctx context.Context, companyID uuid.UUID, cardID uuid.UUID) (*models.Card, error) {
	args := m.Called(ctx, companyID, cardID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Card), args.Error(1)
}

func (m *MockCardRepository) UpdateSpendingLimit(ctx context.Context, id uuid.UUID, spendingLimit int) (*models.Card, error) {
	args := m.Called(ctx, id, spendingLimit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Card), args.Error(1)
}

func TestGetCardByCompanyIDAndCardID(t *testing.T) {
	mockRepo := new(MockCardRepository)
	svc := card.NewService(mockRepo)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		companyID := uuid.New()
		cardID := uuid.New()
		spendingLimitFloat := float64(5000)

		expectedCard := &models.Card{
			ID:            cardID,
			CompanyID:     companyID,
			SpendingLimit: &spendingLimitFloat,
		}

		mockRepo.On("GetCardByCompanyIDAndCardID", ctx, companyID, cardID).Return(expectedCard, nil).Once()

		card, err := svc.GetCardByCompanyIDAndCardID(ctx, companyID, cardID)
		require.NoError(t, err)
		require.NotNil(t, card)

		assert.Equal(t, expectedCard, card)
		mockRepo.AssertExpectations(t)
	})

	t.Run("not_found", func(t *testing.T) {
		companyID := uuid.New()
		cardID := uuid.New()

		mockRepo.On("GetCardByCompanyIDAndCardID", ctx, companyID, cardID).Return(nil, nil).Once()

		card, err := svc.GetCardByCompanyIDAndCardID(ctx, companyID, cardID)
		require.NoError(t, err)
		assert.Nil(t, card)
		mockRepo.AssertExpectations(t)
	})

	t.Run("repository_error", func(t *testing.T) {
		companyID := uuid.New()
		cardID := uuid.New()
		expectedErr := assert.AnError

		mockRepo.On("GetCardByCompanyIDAndCardID", ctx, companyID, cardID).Return(nil, expectedErr).Once()

		card, err := svc.GetCardByCompanyIDAndCardID(ctx, companyID, cardID)
		require.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Nil(t, card)
		mockRepo.AssertExpectations(t)
	})
}

func TestUpdateSpendingLimit(t *testing.T) {
	mockRepo := new(MockCardRepository)
	svc := card.NewService(mockRepo)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		cardID := uuid.New()
		spendingLimit := 5000
		spendingLimitFloat := float64(spendingLimit)

		expectedCard := &models.Card{
			ID:            cardID,
			CompanyID:     uuid.New(),
			SpendingLimit: &spendingLimitFloat,
		}

		mockRepo.On("UpdateSpendingLimit", ctx, cardID, spendingLimit).Return(expectedCard, nil).Once()

		updatedCard, err := svc.UpdateSpendingLimit(ctx, cardID, spendingLimit)
		require.NoError(t, err)
		require.NotNil(t, updatedCard)

		assert.Equal(t, expectedCard, updatedCard)
		assert.Equal(t, spendingLimitFloat, *updatedCard.SpendingLimit)
		mockRepo.AssertExpectations(t)
	})

	t.Run("repository_error", func(t *testing.T) {
		cardID := uuid.New()
		spendingLimit := 5000
		expectedErr := assert.AnError

		mockRepo.On("UpdateSpendingLimit", ctx, cardID, spendingLimit).Return(nil, expectedErr).Once()

		updatedCard, err := svc.UpdateSpendingLimit(ctx, cardID, spendingLimit)
		require.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Nil(t, updatedCard)
		mockRepo.AssertExpectations(t)
	})
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
