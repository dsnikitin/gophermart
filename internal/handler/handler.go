package handler

import (
	"context"

	"github.com/dsnikitin/gophermart/internal/config"
)

type Service interface {
	Register(ctx context.Context, login, password string) error
	Login(ctx context.Context, login, password string) error
}

type Handler struct {
	cfg  *config.Config
	user UserService
}

func New(cfg *config.Config, s Service) *Handler {
	return &Handler{
		cfg:  cfg,
		user: s,
	}
}
