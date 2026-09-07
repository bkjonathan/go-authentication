package middleware

import (
	"slices"
	"strings"

	"github.com/bkjonathan/go-authentication/internal/apperror"
	"github.com/bkjonathan/go-authentication/internal/models"
	"github.com/bkjonathan/go-authentication/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const principalKey = "auth.principal"

// Principal is who the request is from, read from the access token and nothing
// else - no database round trip on an authenticated request.
type Principal struct {
	UserID    uuid.UUID
	Role      models.Role
	SessionID uuid.UUID
}

// Authenticate rejects anything without a valid access token and records who
// the caller is for the handlers behind it.
func (m *Middleware) Authenticate() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		raw, ok := bearerToken(ctx.GetHeader("Authorization"))
		if !ok {
			utils.Fail(ctx, apperror.Unauthorized("missing_token", "Sign in to use this endpoint."))
			return
		}

		claims, err := m.tokens.Parse(raw)
		if err != nil {
			utils.Fail(ctx, apperror.From(err))
			return
		}

		ctx.Set(principalKey, &Principal{
			UserID:    claims.UserID(),
			Role:      claims.Role,
			SessionID: claims.SessionID,
		})
		ctx.Next()
	}
}

// RequireRole guards a route with the roles it accepts. It runs behind
// Authenticate, which is what put the role there.
func (m *Middleware) RequireRole(roles ...models.Role) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		principal, err := CurrentPrincipal(ctx)
		if err != nil {
			utils.Fail(ctx, apperror.From(err))
			return
		}
		if !slices.Contains(roles, principal.Role) {
			utils.Fail(ctx, apperror.Forbidden("forbidden", "Your role does not allow this."))
			return
		}
		ctx.Next()
	}
}

// CurrentPrincipal reads the caller a handler is acting for. The error only
// happens when a route was registered without Authenticate in front of it,
// which is a wiring mistake rather than something a client can cause.
func CurrentPrincipal(ctx *gin.Context) (*Principal, error) {
	principal, ok := ctx.Value(principalKey).(*Principal)
	if !ok {
		return nil, apperror.Unauthorized("missing_token", "Sign in to use this endpoint.")
	}
	return principal, nil
}

func bearerToken(header string) (string, bool) {
	scheme, token, found := strings.Cut(header, " ")
	if !found || !strings.EqualFold(scheme, "Bearer") {
		return "", false
	}
	token = strings.TrimSpace(token)
	return token, token != ""
}
