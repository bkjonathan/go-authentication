package server

import (
	"context"
	"errors"
	"net"
	"net/http"
	"time"

	"github.com/bkjonathan/go-authentication/internal/config"
	"github.com/rs/zerolog"
)

type Server struct {
	http   *http.Server
	logger *zerolog.Logger
}

func New(cfg *config.ServerConfig, handler http.Handler, logger *zerolog.Logger) *Server {
	return &Server{
		http: &http.Server{
			Addr:         net.JoinHostPort("", cfg.Port),
			Handler:      handler,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
		},
		logger: logger,
	}
}

func (s *Server) Start() error {
	s.logger.Info().Str("address", s.http.Addr).Msg("htt server listening")
	if err := s.http.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.http.Shutdown(ctx)
}
