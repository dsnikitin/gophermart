package handler

import (
	"github.com/dsnikitin/gophermart/internal/config"
)

type Service interface {
	UserService
	OrderService
}

type Handler struct {
	cfg   *config.Config
	user  UserService
	order OrderService
}

func New(cfg *config.Config, s Service) *Handler {
	return &Handler{
		cfg:   cfg,
		user:  s,
		order: s,
	}
}
