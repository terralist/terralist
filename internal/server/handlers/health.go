package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Health() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	}
}
