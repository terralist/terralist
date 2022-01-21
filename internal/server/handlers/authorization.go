package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/valentindeaconu/terralist/internal/server/oauth/token"
)

func Authorize() gin.HandlerFunc {
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

		userDetails, err := token.Validate(t)
		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
		}

		c.Set("userName", userDetails.Name)
		c.Set("userEmail", userDetails.Email)

		c.Next()
	}
}
