package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func health(c *gin.Context) {

	c.JSON(http.StatusOK, gin.H{
		"success": "Route healthy",
	})
}