package client

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"ccards/internal/api/request"
	"ccards/internal/api/response"
	"ccards/pkg/errors"
	"ccards/pkg/middleware"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterCompany(c *gin.Context) {
	var req request.RegisterCompany
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.service.RegisterCompany(c.Request.Context(), &req)
	if err != nil {
		switch err {
		case errors.ErrCompanyExists:
			c.JSON(http.StatusConflict, gin.H{"error": "Company with this email already exists"})
		case errors.ErrUserExists:
			c.JSON(http.StatusConflict, gin.H{"error": "User with this email already exists"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register company"})
		}
		return
	}

	c.JSON(http.StatusCreated, resp)
}

func (h *Handler) Login(c *gin.Context) {
	var req request.LoginCompany
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.service.Login(c.Request.Context(), &req)
	if err != nil {
		switch err {
		case errors.ErrInvalidCredentials:
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		case errors.ErrCompanySuspended:
			c.JSON(http.StatusForbidden, gin.H{"error": "Company account is suspended"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to login"})
		}
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *Handler) RefreshToken(c *gin.Context) {
	var req request.RefreshToken
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.service.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *Handler) GetCompany(c *gin.Context) {
	companyID, err := middleware.GetCompanyIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	company, err := h.service.GetCompanyByID(c.Request.Context(), companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve company information"})
		return
	}

	if company == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Company not found"})
		return
	}

	resp := response.Company{
		ID:        company.ID,
		ClientID:  company.ClientID,
		Name:      company.Name,
		Email:     company.Email,
		Address:   company.Address,
		Phone:     company.Phone,
		Status:    company.Status,
		CreatedAt: company.CreatedAt,
		UpdatedAt: company.UpdatedAt,
	}

	c.JSON(http.StatusOK, resp)
}

func (h *Handler) UploadCardCSV(c *gin.Context) {
	companyID, err := middleware.GetCompanyIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get file from request"})
		return
	}
	defer file.Close()

	if !strings.HasSuffix(strings.ToLower(header.Filename), ".csv") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File must be a CSV"})
		return
	}

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read file"})
		return
	}

	if err := h.service.ProcessCardCSVUpload(c.Request.Context(), companyID, fileBytes); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Failed to process CSV: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "CSV uploaded and processed successfully",
	})
}

func (h *Handler) GetCardsToIssue(c *gin.Context) {
	companyID, err := middleware.GetCompanyIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	cards, err := h.service.GetCardsToIssueByClientID(c.Request.Context(), companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve cards to issue"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"cards": cards,
		"count": len(cards),
	})
}
