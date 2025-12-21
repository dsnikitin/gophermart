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
	Accrual AccrualService
}

func New(cfg *config.Config, s *Service) *Handler {
	return &Handler{
		User:    NewUserHandler(cfg, s.User),
		Order:   NewOrderHandler(s.Order, s.Accrual),
		Balance: NewBalanceHandler(s.Balance),
	}
}
