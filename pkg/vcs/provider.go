package vcs

import "github.com/gin-gonic/gin"

type Provider interface {
	GetHeaders() map[string]string
	Authenticate(ctx *gin.Context, body []byte) error
	BuildReleaseEventFromWebhook(body []byte) (*ReleaseEvent, error)
}
