package handler

import (
	"drivo/internal/middleware"
	"drivo/internal/models"
	"drivo/internal/repository"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type RecurringHandler struct {
	recurringRepo *repository.RecurringRepo
}

func NewRecurringHandler(recurringRepo *repository.RecurringRepo) *RecurringHandler {
	return &RecurringHandler{recurringRepo: recurringRepo}
}

func (h *RecurringHandler) Create(c *gin.Context) {
	userID, ok := middleware.GetUserId(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	riderID, _ := uuid.Parse(userID)

	var req struct {
		PickupLat      float64 `json:"pickup_lat" binding:"required"`
		PickupLng      float64 `json:"pickup_lng" binding:"required"`
		DropoffLat     float64 `json:"dropoff_lat" binding:"required"`
		DropoffLng     float64 `json:"dropoff_lng" binding:"required"`
		PickupAddress  string  `json:"pickup_address"`
		DropoffAddress string  `json:"dropoff_address"`
		DaysOfWeek     string  `json:"days_of_week" binding:"required"`
		PickupTime     string  `json:"pickup_time" binding:"required"`
		EndDate        *string `json:"end_date"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	validDays := map[string]bool{"mon": true, "tue": true, "wed": true, "thu": true, "fri": true, "sat": true, "sun": true}
	for _, d := range strings.Split(req.DaysOfWeek, ",") {
		if !validDays[strings.TrimSpace(strings.ToLower(d))] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid day: " + d + ". Use mon,tue,wed,thu,fri,sat,sun"})
			return
		}
	}

	var hour, min int
	if _, err := fmt.Sscanf(req.PickupTime, "%d:%d", &hour, &min); err != nil || hour > 23 || min > 59 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid pickup_time, use HH:MM format e.g. 07:00"})
		return
	}

	ride := &models.RecurringRide{
		RiderID:        riderID,
		PickupLat:      req.PickupLat,
		PickupLng:      req.PickupLng,
		DropoffLat:     req.DropoffLat,
		DropoffLng:     req.DropoffLng,
		PickupAddress:  req.PickupAddress,
		DropoffAddress: req.DropoffAddress,
		DaysOfWeek:     strings.ToLower(req.DaysOfWeek),
		PickupTime:     req.PickupTime,
		Timezone:       "Africa/Lagos",
		IsActive:       true,
		StartDate:      time.Now(),
	}

	if req.EndDate != nil {
		t, err := time.Parse("2006-01-02", *req.EndDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end_date format, use YYYY-MM-DD"})
			return
		}
		ride.EndDate = &t
	}

	if err := h.recurringRepo.Create(c.Request.Context(), ride); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"recurring_ride": ride,
		"message":        "Recurring ride set up successfully! Rides will be auto-booked.",
	})
}

func (h *RecurringHandler) List(c *gin.Context) {
	userID, ok := middleware.GetUserId(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	riderID, _ := uuid.Parse(userID)

	rides, err := h.recurringRepo.GetByRiderID(c.Request.Context(), riderID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"recurring_rides": rides})
}

// DELETE /ride/recurring/:id
func (h *RecurringHandler) Delete(c *gin.Context) {
	userID, ok := middleware.GetUserId(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	riderID, _ := uuid.Parse(userID)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid recurring ride id"})
		return
	}

	if err := h.recurringRepo.Deactivate(c.Request.Context(), id, riderID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Recurring ride cancelled successfully"})
}
