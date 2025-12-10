package models

import (
	"github.com/dsnikitin/gophermart/internal/pkg/validationx"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type AuthRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func (r *AuthRequest) Validate() error {
	return validation.ValidateStruct(r,
		validation.Field(&r.Login, validation.By(func(any) error {
			return validationx.ValidateLogin(r.Login)
		})),
		validation.Field(&r.Password, validation.By(func(any) error {
			return validationx.ValidatePassword(r.Password, r.Login)
		})),
	)
}

type User struct {
	Login    string
	Password string
}
