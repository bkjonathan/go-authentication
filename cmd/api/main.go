package main

import (
	"fmt"

	"github.com/bkjonathan/go-authentication/internal/config"
	"github.com/bkjonathan/go-authentication/internal/config/app"
	"github.com/bkjonathan/go-authentication/internal/logger"
	"github.com/rs/zerolog"
)

func main() {
	log := logger.New()
	if err := run(&log); err != nil {
		log.Fatal().Err(err).Msg("application stopped")
	}
}

func run(log *zerolog.Logger) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	application, err := app.New(cfg, log)
	if err != nil {
		return err
	}
	defer application.Close()
	return application.Run()
}
