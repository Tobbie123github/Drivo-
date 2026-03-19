package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

//  /abc -> only admin can access,  1st level, if the user is authenticated, 2nd level if user is admin
// /xyz ->  any auth user can access, if the user is authenticated
// /bbc ->  anybody can access, no auth needed

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
