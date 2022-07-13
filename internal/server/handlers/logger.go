package handlers

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		t := time.Now()

		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		if raw != "" {
			path = path + "?" + raw
		}

		msg := c.Errors.String()
		if msg == "" {
			msg = "accepted request"
		}

		statusCode := c.Writer.Status()

		var e *zerolog.Event
		switch {
		case statusCode >= 400 && statusCode < 500:
			e = log.Warn()
		case statusCode >= 500:
			e = log.Error()
		default:
			e = log.Info()
		}

		e.Str("method", c.Request.Method).
			Str("path", path).
			Dur("resp_time", time.Since(t)).
			Int("status", statusCode).
			Str("client_ip", c.ClientIP()).
			Msg(msg)
	}
}
