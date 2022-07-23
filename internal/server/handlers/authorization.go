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
		authHeader := c.GetHeader("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		var bearerToken string
		_, err := fmt.Sscanf(authHeader, "Bearer %s", &bearerToken)
		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
		}

		if !strings.HasPrefix(bearerToken, "x-api-key:") {
			userDetails, err := jwt.Extract(bearerToken)
			if err != nil {
				c.AbortWithStatus(http.StatusUnauthorized)
				return
			}

			c.Set("userName", userDetails.Name)
			c.Set("userEmail", userDetails.Email)
		} else {
			var apiKey string
			_, err := fmt.Sscanf(bearerToken, "x-api-key:%s", &apiKey)
			if err != nil {
				c.AbortWithStatus(http.StatusUnauthorized)
			}

			// TODO: Set context from apiKey details
			c.Set("userName", "TODO")
			c.Set("userEmail", "TODO")
			c.Set("authority", "")
		}

		c.Next()
	}
}

func RequireAuthority() gin.HandlerFunc {
	return func(c *gin.Context) {
		if _, ok := c.Get("authority"); !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"errors": []string{"An API key is required to perform this operation."},
			})
			return
		}

		c.Next()
	}
}
