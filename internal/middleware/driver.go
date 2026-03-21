package middleware

import (
	"drivo/internal/models"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RequireActiveDriver(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get(ctxUserIdKey)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		var driver models.Driver
		if err := db.Where("user_id = ?", userID).First(&driver).Error; err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "Driver record not found"})
			c.Abort()
			return
		}

		log.Printf("RequireActiveDriver: driver %s status=%s identity=%v vehicle=%v",
			driver.ID, driver.Status, driver.IsIdentityVerified, driver.IsVehicleVerified)

		if driver.Status != models.DriverActive {
			c.JSON(http.StatusForbidden, gin.H{
				"error": map[models.DriverStatus]string{
					models.DriverPending:   "Your account is pending admin approval",
					models.DriverSuspended: "Your account has been suspended",
					models.DriverBanned:    "Your account has been banned",
				}[driver.Status],
			})
			c.Abort()
			return
		}

		// check status
		if driver.Status != models.DriverActive {
			c.JSON(http.StatusForbidden, gin.H{
				"error":  "Account not active",
				"status": driver.Status,
				"message": map[models.DriverStatus]string{
					models.DriverPending:   "Your account is pending admin approval",
					models.DriverSuspended: "Your account has been suspended",
					models.DriverBanned:    "Your account has been banned",
				}[driver.Status],
			})
			c.Abort()
			return
		}

		// Check identity verified
		if !driver.IsIdentityVerified {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Your identity has not been verified yet",
			})
			c.Abort()
			return
		}

		// Check vehicle verified
		if !driver.IsVehicleVerified {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Your vehicle has not been verified yet",
			})
			c.Abort()
			return
		}

		// check license
		if !driver.LicenseVerified {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Your license has not been verified yet",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
