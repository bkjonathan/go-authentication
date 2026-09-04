package app

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bkjonathan/go-authentication/internal/config"
	"github.com/bkjonathan/go-authentication/internal/server"
	"github.com/rs/zerolog"
)

const shutdownTimeout = 15 * time.Second

type App struct {
	container *container
	server    *server.Server
	logger    *zerolog.Logger
}

func New(cfg *config.Config, logger *zerolog.Logger) (*App, error) {
	c, err := newContainer(cfg, logger)
	if err != nil {
		return nil, err
	}
	router := server.NewRouter(cfg, c.middleware, c.handler)
	return &App{
		container: c,
		server:    server.New(&cfg.Server, router, logger),
		logger:    logger,
	}, nil
}

func (a *App) Run() error {
	serverFailed := make(chan error, 1)
	go func() {
		serverFailed <- a.server.Start()
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverFailed:
		return err
	case sig := <-quit:
		a.logger.Info().Str("signal", sig.String()).Msg("shutting down")
	}
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	return a.server.Shutdown(ctx)
}

func (a *App) Close() {
	if err := a.container.close(); err != nil {
		a.logger.Error().Err(err).Msg("failed to close database")
	}
}
