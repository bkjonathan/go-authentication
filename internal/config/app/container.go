package app

import (
	"fmt"

	"github.com/bkjonathan/go-authentication/internal/config"
	"github.com/bkjonathan/go-authentication/internal/database"
	"github.com/bkjonathan/go-authentication/internal/handlers"
	"github.com/bkjonathan/go-authentication/internal/middleware"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

type container struct {
	db         *gorm.DB
	middleware *middleware.Middleware
	handler    *handlers.Registry
}

func newContainer(cfg *config.Config, logger *zerolog.Logger) (*container, error) {
	db, err := database.New(&cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("connect database %w", err)
	}
	// HTTP layer
	register := &handlers.Registry{}
	return &container{
		db:         db,
		middleware: middleware.New(&cfg.JWT),
		handler:    register,
	}, nil
}

func (c *container) close() error {
	sqlDB, err := c.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
