package app

import (
	"fmt"

	"github.com/bkjonathan/go-authentication/internal/config"
	"github.com/bkjonathan/go-authentication/internal/database"
	"github.com/bkjonathan/go-authentication/internal/handlers"
	"github.com/bkjonathan/go-authentication/internal/middleware"
	"github.com/bkjonathan/go-authentication/internal/repositories"
	"github.com/bkjonathan/go-authentication/internal/services"
	"github.com/bkjonathan/go-authentication/internal/utils"
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

	// Shared plumbing
	tokens := utils.NewTokenIssuer(cfg.JWT.Secret, cfg.JWT.ExpiresIn)
	hasher, err := utils.NewPasswordHasher(cfg.Auth.BcryptCost)
	if err != nil {
		return nil, err
	}

	// Data layer
	store := repositories.NewStore(db)

	// Business layer
	authService := services.NewAuthService(store, tokens, hasher, cfg.JWT.RefreshTokenExpires, cfg.Auth)
	sessionService := services.NewSessionService(store)

	// HTTP layer
	register := &handlers.Registry{
		Auth:     handlers.NewAuthHandler(authService),
		Sessions: handlers.NewSessionHandler(sessionService),
	}

	return &container{
		db:         db,
		middleware: middleware.New(tokens),
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
