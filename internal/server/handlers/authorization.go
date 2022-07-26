package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"terralist/internal/server/services"
	"terralist/pkg/auth/jwt"

	"github.com/gin-gonic/gin"
)

func Authorize(jwt jwt.JWT, apiKeyService services.ApiKeyService) gin.HandlerFunc {
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
				return
			}

			user, err := apiKeyService.GetUserDetails(apiKey)
			if err != nil {
				c.AbortWithStatus(http.StatusUnauthorized)
				return
			}

			c.Set("userName", user.Name)
			c.Set("userEmail", user.Email)
			c.Set("authority", user.AuthorityID)
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
