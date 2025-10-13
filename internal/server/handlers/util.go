package handlers

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// GetFromContext attempts to extract a context key and cast it to a given type.
func GetFromContext[T any](ctx *gin.Context, name string) (*T, error) {
	raw, ok := ctx.Get(name)
	if !ok {
		return nil, fmt.Errorf("key doesn't exist in context")
	}

	res, ok := raw.(*T)
	if !ok {
		return nil, fmt.Errorf("type mismatch")
	}

	return res, nil
}

// MustGetFromContext is a wrapper over the GetFromContext method that swallows the error.
// It should only be used when the key is known to exist in the context.
func MustGetFromContext[T any](ctx *gin.Context, name string) *T {
	r, err := GetFromContext[T](ctx, name)
	if err != nil {
		log.Error().
			Str("name", name).
			Ctx(ctx).
			Err(err).
			Msg("Assumption that a key should exist inside the context was wrong.")
	}

	return r
}
