package middleware

import (
	"drivo/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RequireOnboardingComplete(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		driverID, exists := c.Get(ctxUserIdKey)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		var driver models.Driver
		if err := db.Where("user_id = ?", driverID).First(&driver).Error; err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "Driver record not found"})
			c.Abort()
			return
		}

		if !driver.IsOnboardingCompleted {
			c.JSON(http.StatusForbidden, gin.H{
				"error":           "Onboarding not complete",
				"onboarding_step": driver.OnboardingStep,
				"message":         "Please complete your onboarding before accessing the dashboard",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
