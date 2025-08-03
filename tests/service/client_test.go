package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"ccards/internal/api/request"
	"ccards/internal/client"
	"ccards/pkg/config"
	"ccards/pkg/errors"
	"ccards/pkg/models"
	"ccards/pkg/utils"
	"ccards/tests/setup"
)

// MockRepository is a mock implementation of the client.Repository interface
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) CreateCompany(ctx context.Context, company *models.Company) error {
	args := m.Called(ctx, company)
	return args.Error(0)
}

func (m *MockRepository) GetCompanyByID(ctx context.Context, id uuid.UUID) (*models.Company, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Company), args.Error(1)
}

func (m *MockRepository) GetCompanyByEmail(ctx context.Context, email string) (*models.Company, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Company), args.Error(1)
}

func (m *MockRepository) CreateCardsToIssue(ctx context.Context, cards []*models.CardToIssue) error {
	args := m.Called(ctx, cards)
	return args.Error(0)
}

func (m *MockRepository) GetCardsToIssueByClientID(ctx context.Context, clientID uuid.UUID) ([]*models.CardToIssue, error) {
	args := m.Called(ctx, clientID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.CardToIssue), args.Error(1)
}

func (m *MockRepository) UpdateCardToIssueStatus(ctx context.Context, id uuid.UUID, status string) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

// MockRedis is a mock implementation of the redis.Client
type MockRedis struct {
	mock.Mock
}

func (m *MockRedis) HSet(ctx context.Context, key string, values interface{}) *redis.IntCmd {
	args := m.Called(ctx, key, values)
	return args.Get(0).(*redis.IntCmd)
}

func (m *MockRedis) Expire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd {
	args := m.Called(ctx, key, expiration)
	return args.Get(0).(*redis.BoolCmd)
}

func TestRegisterCompany(t *testing.T) {
	helper := setup.NewTestHelper(t)
	mockRepo := new(MockRepository)
	jwtConfig := config.JWTConfig{
		Secret:               "test-secret",
		AccessTokenDuration:  time.Hour,
		RefreshTokenDuration: time.Hour * 24,
	}

	svc := client.NewService(mockRepo, jwtConfig, helper.Redis)

	t.Run("success", func(t *testing.T) {
		req := &request.RegisterCompany{
			Name:    "Test Company",
			Email:   "test@example.com",
			Address: "123 Test St",
			Phone:   "123-456-7890",
		}

		// Mock the repository calls
		mockRepo.On("GetCompanyByEmail", mock.Anything, req.Email).Return(nil, nil).Once()
		mockRepo.On("CreateCompany", mock.Anything, mock.AnythingOfType("*models.Company")).Return(nil).Once()

		resp, err := svc.RegisterCompany(context.Background(), req)
		require.NoError(t, err)
		require.NotNil(t, resp)

		assert.Equal(t, req.Name, resp.Company.Name)
		assert.Equal(t, req.Email, resp.Company.Email)
		assert.Equal(t, req.Address, resp.Company.Address)
		assert.Equal(t, req.Phone, resp.Company.Phone)
		assert.Equal(t, models.CompanyStatusActive, resp.Company.Status)
		assert.Equal(t, req.Email, resp.Credentials.Email)
		assert.NotEmpty(t, resp.Credentials.Password)

		mockRepo.AssertExpectations(t)
	})

	t.Run("company_exists", func(t *testing.T) {
		req := &request.RegisterCompany{
			Name:    "Existing Company",
			Email:   "existing@example.com",
			Address: "456 Existing St",
			Phone:   "456-789-0123",
		}

		existingCompany := &models.Company{
			ID:    uuid.New(),
			Email: req.Email,
		}

		// Mock the repository calls
		mockRepo.On("GetCompanyByEmail", mock.Anything, req.Email).Return(existingCompany, nil).Once()

		resp, err := svc.RegisterCompany(context.Background(), req)
		require.Error(t, err)
		assert.Equal(t, errors.ErrCompanyExists, err)
		assert.Nil(t, resp)

		mockRepo.AssertExpectations(t)
	})

	t.Run("repository_error", func(t *testing.T) {
		req := &request.RegisterCompany{
			Name:    "Error Company",
			Email:   "error@example.com",
			Address: "789 Error St",
			Phone:   "789-012-3456",
		}

		// Mock the repository calls
		mockRepo.On("GetCompanyByEmail", mock.Anything, req.Email).Return(nil, errors.ErrBadRequest).Once()

		resp, err := svc.RegisterCompany(context.Background(), req)
		require.Error(t, err)
		assert.Equal(t, errors.ErrBadRequest, err)
		assert.Nil(t, resp)

		mockRepo.AssertExpectations(t)
	})
}

func TestLogin(t *testing.T) {
	helper := setup.NewTestHelper(t)
	mockRepo := new(MockRepository)
	jwtConfig := config.JWTConfig{
		Secret:               "test-secret",
		AccessTokenDuration:  time.Hour,
		RefreshTokenDuration: time.Hour * 24,
	}

	svc := client.NewService(mockRepo, jwtConfig, helper.Redis)

	t.Run("success", func(t *testing.T) {
		password := "password123"
		hashedPassword, _ := utils.HashPassword(password)
		companyID := uuid.New()
		company := &models.Company{
			ID:       companyID,
			Name:     "Test Company",
			Email:    "test@example.com",
			Password: hashedPassword,
			Status:   models.CompanyStatusActive,
		}

		req := &request.LoginCompany{
			Email:    company.Email,
			Password: password,
		}

		// Mock the repository calls
		mockRepo.On("GetCompanyByEmail", mock.Anything, req.Email).Return(company, nil).Once()

		resp, err := svc.Login(context.Background(), req)
		require.NoError(t, err)
		require.NotNil(t, resp)

		assert.NotEmpty(t, resp.AccessToken)
		assert.NotEmpty(t, resp.RefreshToken)
		assert.Equal(t, "Bearer", resp.TokenType)
		assert.Equal(t, company.ID, resp.Company.ID)
		assert.Equal(t, company.Name, resp.Company.Name)
		assert.Equal(t, company.Email, resp.Company.Email)

		mockRepo.AssertExpectations(t)
	})

	t.Run("company_not_found", func(t *testing.T) {
		req := &request.LoginCompany{
			Email:    "nonexistent@example.com",
			Password: "password123",
		}

		// Mock the repository calls
		mockRepo.On("GetCompanyByEmail", mock.Anything, req.Email).Return(nil, errors.ErrNotFound).Once()

		resp, err := svc.Login(context.Background(), req)
		require.Error(t, err)
		assert.Equal(t, errors.ErrUnauthorized, err)
		assert.Nil(t, resp)

		mockRepo.AssertExpectations(t)
	})

	t.Run("inactive_company", func(t *testing.T) {
		password := "password123"
		hashedPassword, _ := utils.HashPassword(password)
		company := &models.Company{
			ID:       uuid.New(),
			Name:     "Inactive Company",
			Email:    "inactive@example.com",
			Password: hashedPassword,
			Status:   models.CompanyStatusInactive,
		}

		req := &request.LoginCompany{
			Email:    company.Email,
			Password: password,
		}

		// Mock the repository calls
		mockRepo.On("GetCompanyByEmail", mock.Anything, req.Email).Return(company, nil).Once()

		resp, err := svc.Login(context.Background(), req)
		require.Error(t, err)
		assert.Equal(t, errors.ErrUnauthorized, err)
		assert.Nil(t, resp)

		mockRepo.AssertExpectations(t)
	})

	t.Run("invalid_password", func(t *testing.T) {
		password := "password123"
		hashedPassword, _ := utils.HashPassword(password)
		company := &models.Company{
			ID:       uuid.New(),
			Name:     "Test Company",
			Email:    "test@example.com",
			Password: hashedPassword,
			Status:   models.CompanyStatusActive,
		}

		req := &request.LoginCompany{
			Email:    company.Email,
			Password: "wrongpassword",
		}

		// Mock the repository calls
		mockRepo.On("GetCompanyByEmail", mock.Anything, req.Email).Return(company, nil).Once()

		resp, err := svc.Login(context.Background(), req)
		require.Error(t, err)
		assert.Equal(t, errors.ErrInvalidCredentials, err)
		assert.Nil(t, resp)

		mockRepo.AssertExpectations(t)
	})
}

func TestRefreshToken(t *testing.T) {
	helper := setup.NewTestHelper(t)
	mockRepo := new(MockRepository)
	jwtConfig := config.JWTConfig{
		Secret:               "test-secret",
		AccessTokenDuration:  time.Hour,
		RefreshTokenDuration: time.Hour * 24,
	}

	svc := client.NewService(mockRepo, jwtConfig, helper.Redis)

	t.Run("not_implemented", func(t *testing.T) {
		resp, err := svc.RefreshToken(context.Background(), "some-refresh-token")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not implemented")
		assert.Nil(t, resp)
	})
}

func TestGetCompanyByID(t *testing.T) {
	helper := setup.NewTestHelper(t)
	mockRepo := new(MockRepository)
	jwtConfig := config.JWTConfig{
		Secret:               "test-secret",
		AccessTokenDuration:  time.Hour,
		RefreshTokenDuration: time.Hour * 24,
	}

	svc := client.NewService(mockRepo, jwtConfig, helper.Redis)

	t.Run("success", func(t *testing.T) {
		companyID := uuid.New()
		company := &models.Company{
			ID:     companyID,
			Name:   "Test Company",
			Email:  "test@example.com",
			Status: models.CompanyStatusActive,
		}

		// Mock the repository calls
		mockRepo.On("GetCompanyByID", mock.Anything, companyID).Return(company, nil).Once()

		result, err := svc.GetCompanyByID(context.Background(), companyID)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, company.ID, result.ID)
		assert.Equal(t, company.Name, result.Name)
		assert.Equal(t, company.Email, result.Email)
		assert.Equal(t, company.Status, result.Status)

		mockRepo.AssertExpectations(t)
	})

	t.Run("not_found", func(t *testing.T) {
		companyID := uuid.New()

		// Mock the repository calls
		mockRepo.On("GetCompanyByID", mock.Anything, companyID).Return(nil, nil).Once()

		result, err := svc.GetCompanyByID(context.Background(), companyID)
		require.NoError(t, err)
		assert.Nil(t, result)

		mockRepo.AssertExpectations(t)
	})

	t.Run("repository_error", func(t *testing.T) {
		companyID := uuid.New()

		// Mock the repository calls
		mockRepo.On("GetCompanyByID", mock.Anything, companyID).Return(nil, errors.ErrBadRequest).Once()

		result, err := svc.GetCompanyByID(context.Background(), companyID)
		require.Error(t, err)
		assert.Equal(t, errors.ErrBadRequest, err)
		assert.Nil(t, result)

		mockRepo.AssertExpectations(t)
	})
}

func TestGetCompanyByEmail(t *testing.T) {
	helper := setup.NewTestHelper(t)
	mockRepo := new(MockRepository)
	jwtConfig := config.JWTConfig{
		Secret:               "test-secret",
		AccessTokenDuration:  time.Hour,
		RefreshTokenDuration: time.Hour * 24,
	}

	svc := client.NewService(mockRepo, jwtConfig, helper.Redis)

	t.Run("success", func(t *testing.T) {
		email := "test@example.com"
		company := &models.Company{
			ID:     uuid.New(),
			Name:   "Test Company",
			Email:  email,
			Status: models.CompanyStatusActive,
		}

		// Mock the repository calls
		mockRepo.On("GetCompanyByEmail", mock.Anything, email).Return(company, nil).Once()

		result, err := svc.GetCompanyByEmail(context.Background(), email)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, company.ID, result.ID)
		assert.Equal(t, company.Name, result.Name)
		assert.Equal(t, company.Email, result.Email)
		assert.Equal(t, company.Status, result.Status)

		mockRepo.AssertExpectations(t)
	})

	t.Run("not_found", func(t *testing.T) {
		email := "nonexistent@example.com"

		// Mock the repository calls
		mockRepo.On("GetCompanyByEmail", mock.Anything, email).Return(nil, nil).Once()

		result, err := svc.GetCompanyByEmail(context.Background(), email)
		require.NoError(t, err)
		assert.Nil(t, result)

		mockRepo.AssertExpectations(t)
	})

	t.Run("repository_error", func(t *testing.T) {
		email := "error@example.com"

		// Mock the repository calls
		mockRepo.On("GetCompanyByEmail", mock.Anything, email).Return(nil, errors.ErrBadRequest).Once()

		result, err := svc.GetCompanyByEmail(context.Background(), email)
		require.Error(t, err)
		assert.Equal(t, errors.ErrBadRequest, err)
		assert.Nil(t, result)

		mockRepo.AssertExpectations(t)
	})
}
