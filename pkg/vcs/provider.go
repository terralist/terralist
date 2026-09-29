package vcs

import "github.com/gin-gonic/gin"

type Provider interface {
	// GetHeaders returns the headers to fetch the given urls with. Credentials
	// are included only when every url points to a host of the VCS provider.
	GetHeaders(urls []string) map[string]string
	Authenticate(ctx *gin.Context, body []byte) error
	BuildReleaseEventFromWebhook(body []byte) (*ReleaseEvent, error)
}
