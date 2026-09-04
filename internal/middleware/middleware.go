package middleware

import (
	"net/http"

	"github.com/bkjonathan/go-authentication/internal/config"
	"github.com/gin-gonic/gin"
)

type Middleware struct {
	jwtSecret string
}

func New(jwtCfg *config.JWTConfig) *Middleware {
	return &Middleware{jwtSecret: jwtCfg.Secret}
}

func (m *Middleware) CORS() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Header("Access-Control-Allow-Origin", "*")
		ctx.Header("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		ctx.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if ctx.Request.Method == http.MethodOptions {
			ctx.AbortWithStatus(http.StatusNoContent)
			return
		}

		ctx.Next()
	}
}
