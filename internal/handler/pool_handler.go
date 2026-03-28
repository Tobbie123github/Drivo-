package handler

import (
	"drivo/internal/middleware"
	"drivo/internal/repository"
	"drivo/internal/service"
	"drivo/internal/ws"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type PoolHandler struct {
	poolSvc    *service.PoolService
	riderHub   *ws.RiderHub
	driverRepo *repository.DriverRepo
}

func NewPoolHandler(poolSvc *service.PoolService, riderHub *ws.RiderHub, driverRepo *repository.DriverRepo) *PoolHandler {
	return &PoolHandler{poolSvc: poolSvc, riderHub: riderHub, driverRepo: driverRepo}
}

func (h *PoolHandler) CheckPool(c *gin.Context) {
	userID, ok := middleware.GetUserId(c)

	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
	}

	riderID, _ := uuid.Parse(userID)

	var req struct {
		PickupLat  float64 `json:"pickup_lat" binding:"required"`
		PickupLng  float64 `json:"pickup_lng" binding:"required"`
		DropoffLat float64 `json:"dropoff_lat" binding:"required"`
		DropoffLng float64 `json:"dropoff_lng" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.poolSvc.FindAvailablePools(c.Request.Context(), service.FindPoolInput{
		RiderID:    riderID,
		PickupLat:  req.PickupLat,
		PickupLng:  req.PickupLng,
		DropoffLat: req.DropoffLat,
		DropoffLng: req.DropoffLng,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *PoolHandler) JoinPool(c *gin.Context) {
	userID, ok := middleware.GetUserId(c)

	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
	}
	riderID, _ := uuid.Parse(userID)

	var req struct {
		PoolID      string  `json:"pool_id" binding:"required"`
		PickupLat   float64 `json:"pickup_lat" binding:"required"`
		PickupLng   float64 `json:"pickup_lng" binding:"required"`
		DropoffLat  float64 `json:"dropoff_lat" binding:"required"`
		DropoffLng  float64 `json:"dropoff_lng" binding:"required"`
		PickupAddr  string  `json:"pickup_address"`
		DropoffAddr string  `json:"dropoff_address"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	poolID, err := uuid.Parse(req.PoolID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid pool id"})
		return
	}

	ride, err := h.poolSvc.JoinPool(c.Request.Context(), poolID, service.JoinPoolInput{
		RiderID:     riderID,
		PickupLat:   req.PickupLat,
		PickupLng:   req.PickupLng,
		DropoffLat:  req.DropoffLat,
		DropoffLng:  req.DropoffLng,
		PickupAddr:  req.PickupAddr,
		DropoffAddr: req.DropoffAddr,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ride": ride, "message": "Joined pool successfully"})
}

func (h *PoolHandler) DriverCreatePool(c *gin.Context) {
	userID, ok := middleware.GetUserId(c)

	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
	}
	driverUserID, _ := uuid.Parse(userID)

	driver, err := h.driverRepo.GetDriverByUserID(driverUserID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var req struct {
		PickupLat   float64 `json:"pickup_lat" binding:"required"`
		PickupLng   float64 `json:"pickup_lng" binding:"required"`
		DropoffLat  float64 `json:"dropoff_lat" binding:"required"`
		DropoffLng  float64 `json:"dropoff_lng" binding:"required"`
		PickupAddr  string  `json:"pickup_address"`
		DropoffAddr string  `json:"dropoff_address"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pool, err := h.poolSvc.CreatePool(c.Request.Context(), driver.ID, service.DriverCreatePoolInput{
		DriverID:    driver.ID,
		PickupLat:   req.PickupLat,
		PickupLng:   req.PickupLng,
		DropoffLat:  req.DropoffLat,
		DropoffLng:  req.DropoffLng,
		PickupAddr:  req.PickupAddr,
		DropoffAddr: req.DropoffAddr,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"pool": pool})
}
