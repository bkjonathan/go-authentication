package main

import (
	"fmt"

	"github.com/bkjonathan/go-authentication/internal/config"
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
	_, err := config.Load()
	if err != nil {
		return fmt.Errorf("Load config: %w", err)
	}
	return nil
}
