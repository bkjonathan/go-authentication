package app

import (
	"github.com/bkjonathan/go-authentication/internal/config"
	"github.com/bkjonathan/go-authentication/internal/handlers"
	"github.com/bkjonathan/go-authentication/internal/middleware"
	"github.com/rs/zerolog"
)

type container struct {
	middleware *middleware.Middleware
	handler    *handlers.Registry
}

func newContainer(cfg *config.Config, logger *zerolog.Logger) (*container, error) {

	// HTTP layer
	register := &handlers.Registry{}
	return &container{
		middleware: middleware.New(&cfg.JWT),
		handler:    register,
	}, nil
}

func (c *container) close() error {
	return nil
}
