package server

import (
	"net/http"

	"github.com/bkjonathan/go-authentication/internal/apperror"
	"github.com/bkjonathan/go-authentication/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

// Handler is a route that returns its failure rather than writing one.
type Handler func(*gin.Context) error

// wrap builds the adapter every route is registered through. Errors are
// rendered in exactly one place, and the ones that mean the server is broken -
// rather than the request being wrong - are logged with their cause, which the
// response deliberately does not carry.
func wrap(logger *zerolog.Logger) func(Handler) gin.HandlerFunc {
	return func(handler Handler) gin.HandlerFunc {
		return func(ctx *gin.Context) {
			err := handler(ctx)
			if err == nil {
				return
			}

			appErr := apperror.From(err)
			if appErr.Status >= http.StatusInternalServerError {
				logger.Error().
					Err(err).
					Str("method", ctx.Request.Method).
					Str("path", ctx.FullPath()).
					Msg("request failed")
			}
			utils.Fail(ctx, appErr)
		}
	}
}
