package service

import "github.com/dsnikitin/gophermart/internal/service/usecase"

type Repository interface {
	usecase.UserRepository
}

type Service struct {
	usecase.User
}

func New(repo Repository) *Service {
	return &Service{User: *usecase.NewUser(repo)}
}
