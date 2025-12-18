package handler

import (
	"github.com/dsnikitin/gophermart/internal/config"
)

type Handler struct {
	User    *UserHandler
	Order   *OrderHandler
	Balance *BalanceHandler
}

type Service struct {
	User    UserService
	Order   OrderService
	Balance BalanceService
}

func New(cfg *config.Config, s *Service) *Handler {
	return &Handler{
		User:    NewUser(cfg, s.User),
		Order:   NewOrder(s.Order),
		Balance: NewBalance(s.Balance),
	}
}
