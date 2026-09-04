package server

import (
	"net/http"

	"github.com/bkjonathan/go-authentication/internal/config"
	"github.com/bkjonathan/go-authentication/internal/handlers"
	"github.com/bkjonathan/go-authentication/internal/middleware"
	"github.com/gin-gonic/gin"
)

func NewRouter(cfg *config.Config, mw *middleware.Middleware, h *handlers.Registry) *gin.Engine {
	gin.SetMode(cfg.Server.GinMode)

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery(), mw.CORS())
	router.GET("/health", healthCheck)

	return router

}

// healthCheck answers the load balancer, outside /api/v1 and outside auth.
func healthCheck(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
}
