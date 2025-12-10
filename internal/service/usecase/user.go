package usecase

import (
	"context"

	"github.com/dsnikitin/gophermart/internal/models"
	"github.com/dsnikitin/gophermart/internal/pkg/errx"

	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

type UserRepository interface {
	Create(ctx context.Context, login, password string) error
	Get(ctx context.Context, login string) (models.User, error)
}

type User struct {
	r UserRepository
}

func NewUser(r UserRepository) *User {
	return &User{r: r}
}

func (u *User) Register(ctx context.Context, login, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return errors.Wrap(err, "hash password")
	}

	err = u.r.Create(ctx, login, string(hash))
	return errors.Wrap(err, "repo create user")
}

func (u *User) Login(ctx context.Context, login, password string) error {
	user, err := u.r.Get(ctx, login)
	if err != nil {
		return errors.Wrap(err, "repo get user")
	}

	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		if !errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return errors.Wrap(err, "compare hash and password")
		}

		return errx.ErrInvalidPassword
	}

	return nil
}
