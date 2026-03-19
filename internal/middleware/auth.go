package middleware

import (
	"drivo/internal/auth"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// store auth data info into gin context

const (
	ctxUserIdKey = "auth.userId"
	ctxRoleKey   = "auth.role"
)

func AuthRequired(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := strings.TrimSpace(c.GetHeader("Authorization"))

		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Missing Unauthorized Token",
			})

			return
		}

		parts := strings.SplitN(authHeader, " ", 2)

		if len(parts) != 2 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid Authorization format",
			})
			return
		}

		scheme := strings.TrimSpace(parts[0])
		tokenString := strings.TrimSpace(parts[1])

		if !strings.EqualFold(scheme, "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization must be Bearer format",
			})
			return
		}

		if tokenString == " " {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Missiong token",
			})
			return
		}

		claims, err := auth.VerifyToken(jwtSecret, tokenString)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired token",
			})
			return
		}

		c.Set(ctxUserIdKey, claims.Subject)
		c.Set(ctxRoleKey, claims.Role)

		c.Next()

	}

}

func GetUserId(c *gin.Context) (string, bool) {

	res, ok := c.Get(ctxUserIdKey)

	if !ok {
		return "", false
	}

	userId, ok := res.(string)

	return userId, ok
}

func GetRole(c *gin.Context) (string, bool) {

	res, ok := c.Get(ctxRoleKey)

	if !ok {
		return "", false
	}

	role, ok := res.(string)

	return role, ok
}
