package service

import (
	"context"

	"github.com/dsnikitin/gophermart/internal/models"
	"github.com/dsnikitin/gophermart/internal/pkg/errx"

	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user models.User) error
	GetUser(ctx context.Context, login string) (models.User, error)
}

type UserService struct {
	r UserRepository
}

func NewUser(r UserRepository) *UserService {
	return &UserService{r: r}
}

func (u *UserService) Register(ctx context.Context, req models.RegisterRequest) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return errors.Wrap(err, "hash password")
	}

	req.Password = string(hash)

	err = u.r.CreateUser(ctx, req.User)
	return errors.Wrap(err, "create user")
}

func (u *UserService) Login(ctx context.Context, req models.LoginRequest) error {
	user, err := u.r.GetUser(ctx, req.Login)
	if err != nil {
		return errors.Wrap(err, "get user")
	}

	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		if !errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return errors.Wrap(err, "compare hash and password")
		}

		return errx.ErrInvalidPassword
	}

	return nil
}
