package server

import (
	"net/http"
	"time"

	"github.com/dsnikitin/gophermart/internal/config"
	"github.com/dsnikitin/gophermart/internal/handler"
)

type Server struct {
	*http.Server
}

func New(cfg *config.Config, h *handler.Handler) *Server {
	return &Server{
		Server: &http.Server{
			Addr:         cfg.ServerAddr,
			Handler:      initRouter(cfg, h),
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  30 * time.Second,
		},
	}
}
