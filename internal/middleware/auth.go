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

// func AuthRequired(jwtSecret string) gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		var tokenString string

// 		if qToken := c.Query("token"); qToken != "" {
// 			tokenString = qToken
// 		} else {

// 			authHeader := strings.TrimSpace(c.GetHeader("Authorization"))

// 			if authHeader == "" {
// 				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
// 					"error": "Missing Unauthorized Token",
// 				})
// 				return
// 			}

// 			parts := strings.SplitN(authHeader, " ", 2)

// 			if len(parts) != 2 {
// 				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
// 					"error": "Invalid Authorization format",
// 				})
// 				return
// 			}

// 			scheme := strings.TrimSpace(parts[0])

// 			if !strings.EqualFold(scheme, "Bearer") {
// 				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
// 					"error": "Authorization must be Bearer format",
// 				})
// 				return
// 			}

// 			tokenString = strings.TrimSpace(parts[1])

// 			if tokenString == "" {
// 				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
// 					"error": "Missing token",
// 				})
// 				return
// 			}
// 		}

// 		claims, err := auth.VerifyToken(jwtSecret, tokenString)

// 		if err != nil {
// 			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
// 				"error": "Invalid or expired token",
// 			})
// 			return
// 		}

// 		c.Set(ctxUserIdKey, claims.Subject)
// 		c.Set(ctxRoleKey, claims.Role)

// 		c.Next()
// 	}
// }

func AuthRequired(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var tokenString string

		authHeader := strings.TrimSpace(c.GetHeader("Authorization"))
		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			tokenString = strings.TrimPrefix(authHeader, "Bearer ")
		}

		if tokenString == "" {
			tokenString = c.Query("token")
		}

		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing token"})
			return
		}

		claims, err := auth.VerifyToken(jwtSecret, tokenString)
		if err != nil {
			// ← Check if this is a WebSocket upgrade request
			if c.GetHeader("Upgrade") == "websocket" {
				// Return 401 — browser WebSocket will close with code 1006
				c.AbortWithStatus(http.StatusUnauthorized)
				return
			}
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
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
