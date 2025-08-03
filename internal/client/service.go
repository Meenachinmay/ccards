package client

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"ccards/internal/api/request"
	"ccards/internal/api/response"
	"ccards/pkg/config"
	"ccards/pkg/errors"
	"ccards/pkg/models"
	"ccards/pkg/utils"
)

type service struct {
	repo      Repository
	jwtConfig config.JWTConfig
	redis     *redis.Client
}

func NewService(repo Repository, jwtConfig config.JWTConfig, redis *redis.Client) Service {
	return &service{
		repo:      repo,
		jwtConfig: jwtConfig,
		redis:     redis,
	}
}

func (s *service) RegisterCompany(ctx context.Context, req *request.RegisterCompany) (*response.RegisterCompanyResponse, error) {
	existingCompany, err := s.repo.GetCompanyByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if existingCompany != nil {
		return nil, errors.ErrCompanyExists
	}

	password := utils.GenerateRandomPassword(12)
	passwordHash, err := utils.HashPassword(password)
	if err != nil {
		return nil, err
	}

	company := &models.Company{
		ID:       uuid.New(),
		Name:     req.Name,
		Email:    req.Email,
		Password: passwordHash,
		Address:  req.Address,
		Phone:    req.Phone,
		Status:   models.CompanyStatusActive,
	}

	if err := s.repo.CreateCompany(ctx, company); err != nil {
		return nil, err
	}

	resp := &response.RegisterCompanyResponse{
		Company: response.Company{
			ID:        company.ID,
			ClientID:  company.ClientID,
			Name:      company.Name,
			Email:     company.Email,
			Address:   company.Address,
			Phone:     company.Phone,
			Status:    company.Status,
			CreatedAt: company.CreatedAt,
			UpdatedAt: company.UpdatedAt,
		},
	}

	resp.Credentials.Email = req.Email
	resp.Credentials.Password = password

	return resp, nil
}

func (s *service) Login(ctx context.Context, req *request.LoginCompany) (*response.LoginResponse, error) {
	companyFound, err := s.repo.GetCompanyByEmail(ctx, req.Email)
	if err != nil {
		return nil, errors.ErrUnauthorized // throw some error here
	}

	if companyFound.Status != models.CompanyStatusActive {
		return nil, errors.ErrUnauthorized
	}

	if err := utils.ComparePassword(companyFound.Password, req.Password); err != nil {
		return nil, errors.ErrInvalidCredentials
	}

	accessToken, err := utils.GenerateJWT(companyFound.ID, s.jwtConfig.Secret, s.jwtConfig.AccessTokenDuration)
	if err != nil {
		return nil, errors.ErrBadRequest
	}

	refreshToken, err := utils.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	sessionKey := fmt.Sprintf("session:%s", companyFound.ID.String())
	sessionData := map[string]interface{}{
		"company_id":    companyFound.ID.String(),
		"refresh_token": refreshToken,
	}

	if err := s.redis.HSet(ctx, sessionKey, sessionData).Err(); err != nil {
		return nil, err
	}
	if err := s.redis.Expire(ctx, sessionKey, s.jwtConfig.RefreshTokenDuration).Err(); err != nil {
		return nil, err
	}

	return &response.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(s.jwtConfig.AccessTokenDuration),
		TokenType:    "Bearer",
		Company: response.Company{
			ID:        companyFound.ID,
			Name:      companyFound.Name,
			Email:     companyFound.Email,
			Address:   companyFound.Address,
			Phone:     companyFound.Phone,
			Status:    companyFound.Status,
			CreatedAt: companyFound.CreatedAt,
			UpdatedAt: companyFound.UpdatedAt,
		},
	}, nil
}

func (s *service) RefreshToken(ctx context.Context, refreshToken string) (*response.RefreshTokenResponse, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *service) GetCompanyByID(ctx context.Context, id uuid.UUID) (*models.Company, error) {
	return s.repo.GetCompanyByID(ctx, id)
}

func (s *service) GetCompanyByEmail(ctx context.Context, email string) (*models.Company, error) {
	return s.repo.GetCompanyByEmail(ctx, email)
}

func (s *service) ProcessCardCSVUpload(ctx context.Context, clientID uuid.UUID, csvData []byte) error {
	reader := csv.NewReader(bytes.NewReader(csvData))

	header, err := reader.Read()
	if err != nil {
		return fmt.Errorf("failed to read CSV header: %w", err)
	}

	expectedHeaders := []string{"employee_id", "employee_email"}
	if len(header) < len(expectedHeaders) {
		return fmt.Errorf("invalid CSV format: expected at least %d columns", len(expectedHeaders))
	}

	for i, expectedHeader := range expectedHeaders {
		if header[i] != expectedHeader {
			return fmt.Errorf("invalid CSV format: expected header %q at position %d, got %q", expectedHeader, i, header[i])
		}
	}

	var cardsToIssue []*models.CardToIssue
	lineNumber := 1

	for {
		lineNumber++
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error reading line %d: %w", lineNumber, err)
		}

		if len(record) < 2 {
			return fmt.Errorf("invalid data at line %d: expected at least 2 columns", lineNumber)
		}

		employeeID, err := uuid.Parse(record[0])
		if err != nil {
			return fmt.Errorf("invalid employee ID at line %d: %w", lineNumber, err)
		}

		employeeEmail := record[1]
		if employeeEmail == "" {
			return fmt.Errorf("empty employee email at line %d", lineNumber)
		}

		cardToIssue := &models.CardToIssue{
			ID:            uuid.New(),
			ClientID:      clientID,
			CardID:        uuid.New(),
			EmployeeID:    employeeID,
			EmployeeEmail: employeeEmail,
			Status:        models.CardToIssueStatusPending,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}

		cardsToIssue = append(cardsToIssue, cardToIssue)
	}

	if len(cardsToIssue) == 0 {
		return fmt.Errorf("no valid records found in CSV")
	}

	if err := s.repo.CreateCardsToIssue(ctx, cardsToIssue); err != nil {
		return fmt.Errorf("failed to create cards to issue: %w", err)
	}

	return nil
}

func (s *service) GetCardsToIssueByClientID(ctx context.Context, clientID uuid.UUID) ([]*models.CardToIssue, error) {
	return s.repo.GetCardsToIssueByClientID(ctx, clientID)
}
