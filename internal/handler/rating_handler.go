package handler

import (
	"drivo/internal/middleware"
	"drivo/internal/models"
	"drivo/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type RatingHandler struct {
	ratingSvc *service.RatingService
}

func NewRatingHandler(ratingSvc *service.RatingService) *RatingHandler {
	return &RatingHandler{ratingSvc: ratingSvc}
}

func (h *RatingHandler) RateDriver(c *gin.Context) {

	userID, ok := middleware.GetUserId(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	riderID, _ := uuid.Parse(userID)

	var input models.RatingInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if err := h.ratingSvc.RateDriver(c.Request.Context(), riderID, input); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Driver rated successfully"})

}
