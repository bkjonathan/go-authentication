package app

import (
	"github.com/bkjonathan/go-authentication/internal/config"
	"github.com/rs/zerolog"
)

type container struct {
}

func newContainer(cfg *config.Config, logger *zerolog.Logger) (*container, error) {

	return &container{}, nil
}
