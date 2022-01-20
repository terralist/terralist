package middlewares

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func AuditLogging() gin.HandlerFunc {
	return func(c *gin.Context) {
		logrus.WithFields(logrus.Fields{
			"user_name":      c.GetString("userName"),
			"user_email":     c.GetString("userEmail"),
			"request_uri":    c.Request.RequestURI,
			"request_method": c.Request.Method,
		}).Info("request allowed")

		c.Next()
	}
}
