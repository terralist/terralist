package controllers

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

func Authorize(c *gin.Context) {
	// clientId := c.Query("client_id")
	// codeChallenge := c.Query("code_challenge")
	// codeChallengeMethod := c.Query("code_challenge_method")
	// redirectUri := c.Query("redirect_uri")
	// responseType := c.Query("response_type")
	// state := c.Query("state")
}

func TokenValidate(c *gin.Context) {
	fmt.Printf("TokenValidate: c.Request: %v\n", c.Request)
}
