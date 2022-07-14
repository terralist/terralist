package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"terralist/pkg/auth/jwt"

	"github.com/gin-gonic/gin"
)

func Authorize(jwt jwt.JWT) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")

		if !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		var t string
		_, err := fmt.Sscanf(header, "Bearer %s", &t)

		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
		}

		userDetails, err := jwt.Extract(t)
		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
		} else {
			c.Set("userName", userDetails.Name)
			c.Set("userEmail", userDetails.Email)
		}
		
		c.Next()
	}
}
