package server

import (
	"net/http"

	"github.com/bkjonathan/go-authentication/internal/config"
	"github.com/bkjonathan/go-authentication/internal/handlers"
	"github.com/bkjonathan/go-authentication/internal/middleware"
	"github.com/bkjonathan/go-authentication/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

func NewRouter(cfg *config.Config, mw *middleware.Middleware, h *handlers.Registry, logger *zerolog.Logger) *gin.Engine {
	gin.SetMode(cfg.Server.GinMode)
	utils.UseJSONFieldNames()

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery(), mw.CORS())
	router.GET("/health", healthCheck)

	handle := wrap(logger)
	v1 := router.Group("/api/v1")
	registerAuthRoutes(v1, mw, h.Auth, handle)
	registerSessionRoutes(v1, mw, h.Sessions, handle)

	return router
}

func registerAuthRoutes(group *gin.RouterGroup, mw *middleware.Middleware, h *handlers.AuthHandler, handle func(Handler) gin.HandlerFunc) {
	auth := group.Group("/auth")
	auth.POST("/register", handle(h.Register))
	auth.POST("/login", handle(h.Login))
	auth.POST("/refresh", handle(h.Refresh))
	auth.POST("/logout", handle(h.Logout))

	// Behind the access token: everything that acts on the account the caller
	// is already signed in to.
	signedIn := auth.Group("", mw.Authenticate())
	signedIn.GET("/me", handle(h.Me))
	signedIn.POST("/change-password", handle(h.ChangePassword))
}

func registerSessionRoutes(group *gin.RouterGroup, mw *middleware.Middleware, h *handlers.SessionHandler, handle func(Handler) gin.HandlerFunc) {
	sessions := group.Group("/sessions", mw.Authenticate())
	sessions.GET("", handle(h.List))
	sessions.DELETE("", handle(h.RevokeOthers))
	sessions.DELETE("/:id", handle(h.Revoke))
}

// healthCheck answers the load balancer, outside /api/v1 and outside auth.
func healthCheck(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
}
