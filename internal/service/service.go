package service

import "github.com/dsnikitin/gophermart/internal/service/usecase"

type Repository interface {
	usecase.UserRepository
	usecase.OrderRepository
}

type Service struct {
	*usecase.User
	*usecase.Order
}

func New(repo Repository) *Service {
	return &Service{
		User:  usecase.NewUser(repo),
		Order: usecase.NewOrder(repo),
	}
}
