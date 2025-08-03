package card

import (
	"ccards/internal/api/request"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"ccards/pkg/middleware"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) GetCards(c *gin.Context) {
	companyID, err := middleware.GetCompanyIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	cards, err := h.service.GetCardsByCompanyID(c.Request.Context(), companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve cards"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"cards": cards,
		"count": len(cards),
	})
}

func (h *Handler) UpdateSpendingLimit(c *gin.Context) {
	var req request.CardSetSpendingLimit

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// get the company
	companyID, err := middleware.GetCompanyIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	cardID := c.Query("cardId")
	cardUUID, err := uuid.Parse(cardID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid card ID format"})
		return
	}
	card, err := h.service.GetCardByCompanyIDAndCardID(c, companyID, cardUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve card"})
		return
	}
	if card == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Card not found"})
		return
	}

	updatedCard, err := h.service.UpdateSpendingLimit(c, card.ID, req.SpendingLimit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update card"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"card": updatedCard,
	})
}
