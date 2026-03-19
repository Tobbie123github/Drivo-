package handler

import (
	"drivo/internal/middleware"
	"drivo/internal/models"
	"drivo/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type RideHandler struct {
	rideSvc *service.RideService
}

func NewRideHandler(rideSvc *service.RideService) *RideHandler {
	return &RideHandler{rideSvc: rideSvc}
}

func (h *RideHandler) RequestRide(c *gin.Context) {
	userID, ok := middleware.GetUserId(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	riderID, _ := uuid.Parse(userID)

	var input models.RideRequestInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	ride, err := h.rideSvc.RequestRide(c.Request.Context(), riderID, input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":        "Ride requested, finding your driver",
		"ride_id":        ride.ID,
		"estimated_fare": ride.EstimatedFare,
		"distance_km":    ride.DistanceKm,
	})
}