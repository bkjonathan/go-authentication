// Package handlers is the thin edge of the API: bind the request, delegate,
// return. A handler returns its error rather than writing one, so every failure
// is rendered in one place.
package handlers

import (
	"github.com/bkjonathan/go-authentication/internal/dto"
	"github.com/gin-gonic/gin"
)

type Registry struct {
	Auth     *AuthHandler
	Sessions *SessionHandler
}

// clientInfo is what the request itself reveals about the device that made it.
// ClientIP honours the proxy headers gin is configured to trust, so it is the
// real caller rather than the load balancer.
func clientInfo(ctx *gin.Context) dto.ClientInfo {
	return dto.ClientInfo{
		IPAddress: ctx.ClientIP(),
		UserAgent: ctx.Request.UserAgent(),
	}
}
