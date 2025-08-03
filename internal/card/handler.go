package card

import (
	"net/http"

	"github.com/gin-gonic/gin"

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
