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

	if ride.IsScheduled {
		c.JSON(http.StatusOK, gin.H{
			"message":        "Ride scheduled successfully",
			"ride_id":        ride.ID,
			"scheduled_time": ride.ScheduledAt,
			"estimated_fare": ride.EstimatedFare,
			"distance_km":    ride.DistanceKm,
		})
		return
	} else {
		c.JSON(http.StatusCreated, gin.H{
			"message":        "Ride requested, finding your driver",
			"ride_id":        ride.ID,
			"estimated_fare": ride.EstimatedFare,
			"distance_km":    ride.DistanceKm,
		})
		return
	}

}

func (h *RideHandler) CancelRide(c *gin.Context) {

	userID, ok := middleware.GetUserId(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	riderID, _ := uuid.Parse(userID)

	var input struct {
		RideID string `json:"ride_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	rideID, err := uuid.Parse(input.RideID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ride id"})
		return
	}

	if err := h.rideSvc.CancelRide(c.Request.Context(), riderID, rideID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Ride cancelled"})
}

func (h *RideHandler) DriverCancelRide(c *gin.Context) {
	userID, ok := middleware.GetUserId(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	driverUserID, _ := uuid.Parse(userID)

	var input struct {
		RideID string `json:"ride_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	rideID, err := uuid.Parse(input.RideID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ride id"})
		return
	}

	if err := h.rideSvc.DriverCancelRide(c.Request.Context(), driverUserID, rideID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Ride cancelled"})
}

func (h *RideHandler) GetRiderHistory(c *gin.Context) {
	userID, ok := middleware.GetUserId(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	riderID, _ := uuid.Parse(userID)

	rides, err := h.rideSvc.GetRiderHistory(c.Request.Context(), riderID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"rides": rides,
		"total": len(rides),
	})
}

func (h *RideHandler) GetDriverHistory(c *gin.Context) {
	userID, ok := middleware.GetUserId(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	driverUserID, _ := uuid.Parse(userID)

	rides, err := h.rideSvc.GetDriverHistory(c.Request.Context(), driverUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"rides": rides,
		"total": len(rides),
	})
}
