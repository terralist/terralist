package middlewares

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/valentindeaconu/terralist/utils"
)

func Authorize() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")

		if !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		var token string
		_, err := fmt.Sscanf(header, "Bearer %s", &token)

		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
		}

		err = utils.Validate(token)
		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
		}

		c.Next()
	}
}
