package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)


func RequireDriver() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, ok := GetRole(c)

		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "UnAuthorized",
			})
			return
		}

		if !strings.EqualFold(role, "driver") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Only Driver can have access",
			})
			return
		}

		c.Next()
	}
}

func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, ok := GetRole(c)

		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "UnAuthorized",
			})
			return
		}

		if !strings.EqualFold(role, "admin") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Only Admin can have access",
			})
			return
		}

		c.Next()
	}
}
