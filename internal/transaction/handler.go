package transaction

import (
	"ccards/pkg/models"
	"net/http"

	"ccards/internal/api/request"
	"ccards/internal/api/response"
	"ccards/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Pay(c *gin.Context) {
	var req request.Transaction

	if txReq, exists := c.Get("transaction_request"); exists {
		txReqPtr, ok := txReq.(*request.Transaction)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Invalid transaction request format in context",
			})
			return
		}
		req = *txReqPtr
	} else {
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid request body",
				"details": err.Error(),
			})
			return
		}
	}

	companyID, exists := c.Get("company_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Company ID not found in context",
		})
		return
	}

	companyUUID, ok := companyID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid company ID type",
		})
		return
	}

	if req.CompanyID != companyUUID {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Company ID mismatch",
		})
		return
	}

	card, exists := c.Get("card")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Card information not found",
		})
		return
	}

	cardModel, ok := card.(*models.Card)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Invalid card information format",
		})
		return
	}
	cardLastFour := cardModel.LastFour

	transaction, remainingBalance, err := h.service.ProcessPayment(
		c.Request.Context(),
		req.CompanyID,
		req.CardID,
		req.Amount,
		req.MerchantCategory,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to process payment",
			"details": err.Error(),
		})
		return
	}

	resp := response.TransactionResponse{
		Transaction: response.Transaction{
			ID:               transaction.ID,
			CardID:           transaction.CardID,
			CompanyID:        transaction.CompanyID,
			TransactionType:  transaction.TransactionType,
			Amount:           transaction.Amount,
			MerchantName:     transaction.MerchantName,
			MerchantCategory: transaction.MerchantCategory,
			Description:      transaction.Description,
			Status:           transaction.Status,
			ProcessedAt:      transaction.ProcessedAt,
			CreatedAt:        transaction.CreatedAt,
			UpdatedAt:        transaction.UpdatedAt,
		},
		RemainingBalance: remainingBalance,
		CardLastFour:     cardLastFour,
	}

	c.JSON(http.StatusOK, resp)
}

func (h *Handler) GetTransactionHistory(c *gin.Context) {
	companyID, exists := c.Get("company_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Company ID not found in context",
		})
		return
	}

	companyUUID, ok := companyID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid company ID type",
		})
		return
	}

	var err error

	// Get pagination parameters
	page := utils.GetIntParam(c, "page", 1)
	pageSize := utils.GetIntParam(c, "page_size", 20)
	offset := (page - 1) * pageSize

	// Get card ID if provided (optional)
	cardIDStr := c.Query("card_id")

	var transactions []*models.Transaction
	if cardIDStr != "" {
		cardID, err := uuid.Parse(cardIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid card ID",
			})
			return
		}

		transactions, err = h.service.GetCardTransactions(c.Request.Context(), cardID, pageSize, offset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to get transactions",
				"details": err.Error(),
			})
			return
		}
	} else {
		transactions, err = h.service.GetCompanyTransactions(c.Request.Context(), companyUUID, pageSize, offset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to get transactions",
				"details": err.Error(),
			})
			return
		}
	}

	respTransactions := make([]response.Transaction, len(transactions))
	for i, tx := range transactions {
		respTransactions[i] = response.Transaction{
			ID:               tx.ID,
			CardID:           tx.CardID,
			CompanyID:        tx.CompanyID,
			TransactionType:  tx.TransactionType,
			Amount:           tx.Amount,
			MerchantName:     tx.MerchantName,
			MerchantCategory: tx.MerchantCategory,
			Description:      tx.Description,
			Status:           tx.Status,
			ProcessedAt:      tx.ProcessedAt,
			CreatedAt:        tx.CreatedAt,
			UpdatedAt:        tx.UpdatedAt,
		}
	}

	resp := response.TransactionListResponse{
		Transactions: respTransactions,
		Total:        len(respTransactions),
		Page:         page,
		PageSize:     pageSize,
	}

	c.JSON(http.StatusOK, resp)
}

func (h *Handler) GetTransaction(c *gin.Context) {
	transactionIDStr := c.Param("id")
	transactionID, err := uuid.Parse(transactionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid transaction ID",
		})
		return
	}

	transaction, err := h.service.GetTransaction(c.Request.Context(), transactionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Transaction not found",
		})
		return
	}

	companyID, _ := c.Get("company_id")
	companyUUID, _ := companyID.(uuid.UUID)

	if transaction.CompanyID != companyUUID {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Access denied",
		})
		return
	}

	resp := response.Transaction{
		ID:               transaction.ID,
		CardID:           transaction.CardID,
		CompanyID:        transaction.CompanyID,
		TransactionType:  transaction.TransactionType,
		Amount:           transaction.Amount,
		MerchantName:     transaction.MerchantName,
		MerchantCategory: transaction.MerchantCategory,
		Description:      transaction.Description,
		Status:           transaction.Status,
		ProcessedAt:      transaction.ProcessedAt,
		CreatedAt:        transaction.CreatedAt,
		UpdatedAt:        transaction.UpdatedAt,
	}

	c.JSON(http.StatusOK, resp)
}
