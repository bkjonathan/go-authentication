package app

import (
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
	server *server.Server
	logger *zerolog.Logger
}

func New(cfg *config.Config, logger *zerolog.Logger) (*App, error) {
	return &App{
		server: server.New(&cfg.Server, router),
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
		a.logge
	}

	return nil
}
