package handler

import (
	"drivo/internal/jobs"
	"drivo/internal/service"
	"drivo/internal/workers"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AdminHandler struct {
	adminSvc *service.AdminService
}

func NewAdminHandler(adminSvc *service.AdminService) *AdminHandler {
	return &AdminHandler{adminSvc: adminSvc}
}

func (h *AdminHandler) GetDashboardStats(c *gin.Context) {
	stats, err := h.adminSvc.GetDashboardStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"stats": stats})
}

func (h *AdminHandler) VerifyDriverIdentity(c *gin.Context) {
	driverID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid driver id"})
		return
	}

	if err := h.adminSvc.VerifyDriverIdentity(c.Request.Context(), driverID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Driver identity verified"})
}

func (h *AdminHandler) VerifyDriverVehicle(c *gin.Context) {
	driverID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid driver id"})
		return
	}

	if err := h.adminSvc.VerifyDriverVehicle(c.Request.Context(), driverID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Driver vehicle verified"})
}

func (h *AdminHandler) VerifyDriverLicense(c *gin.Context) {
	driverID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid driver id"})
		return
	}

	if err := h.adminSvc.VerifyDriverLicense(c.Request.Context(), driverID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Driver license verified"})
}

func (h *AdminHandler) GetAllDrivers(c *gin.Context) {

	status := c.Query("status")

	drivers, err := h.adminSvc.GetAllDrivers(c.Request.Context(), status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"drivers": drivers,
		"total":   len(drivers),
	})
}

func (h *AdminHandler) ApproveDriver(c *gin.Context) {
	driverID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid driver id"})
		return
	}

	d, err := h.adminSvc.ApproveDriver(c.Request.Context(), driverID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if d.User.Email != nil {
		workers.EmailQueue <- jobs.EmailJob{
			Type: jobs.EmailTypeDriverApproved,
			To:   *d.User.Email,
			Name: d.User.Name,
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Driver approved"})
}

func (h *AdminHandler) RejectDriver(c *gin.Context) {
	driverID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid driver id"})
		return
	}

	if err := h.adminSvc.RejectDriver(c.Request.Context(), driverID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Driver rejected"})
}

func (h *AdminHandler) SuspendDriver(c *gin.Context) {
	driverID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid driver id"})
		return
	}

	if err := h.adminSvc.SuspendDriver(c.Request.Context(), driverID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Driver suspended"})
}

func (h *AdminHandler) BanDriver(c *gin.Context) {
	driverID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid driver id"})
		return
	}

	if err := h.adminSvc.BanDriver(c.Request.Context(), driverID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Driver banned"})
}

func (h *AdminHandler) GetAllRiders(c *gin.Context) {
	riders, err := h.adminSvc.GetAllRiders(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"riders": riders,
		"total":  len(riders),
	})
}

func (h *AdminHandler) GetAllRides(c *gin.Context) {

	status := c.Query("status")

	rides, err := h.adminSvc.GetAllRides(c.Request.Context(), status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"rides": rides,
		"total": len(rides),
	})
}
