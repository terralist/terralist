package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func AuditLogging(logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger.WithFields(logrus.Fields{
			"user_name":      c.GetString("userName"),
			"user_email":     c.GetString("userEmail"),
			"request_uri":    c.Request.RequestURI,
			"request_method": c.Request.Method,
		}).Info("request allowed")

		c.Next()
	}
}
