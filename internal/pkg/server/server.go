package server

import (
	"context"
	"net/http"
	"time"

	"github.com/dsnikitin/gophermart/internal/config"

	"github.com/pkg/errors"
)

type HTTPServer struct {
	*http.Server
}

func NewHTTPServer(cfg *config.Config, h http.Handler) *HTTPServer {
	return &HTTPServer{
		Server: &http.Server{
			Addr:         cfg.ServerAddr,
			Handler:      h,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  30 * time.Second,
		},
	}
}

func (s *HTTPServer) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := s.Shutdown(ctx)
	return errors.Wrap(err, "server shutdown")
}
