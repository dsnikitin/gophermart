package models

import (
	"github.com/dsnikitin/gophermart/internal/pkg/validationx"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type LoginKey struct{}

type User struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func (m *User) ScanFields() []any {
	return []any{&m.Login, &m.Password}
}

type RegisterRequest struct {
	User
}

func (m RegisterRequest) Validate() error {
	return validation.ValidateStruct(&m,
		validation.Field(&m.Login, validation.By(func(any) error {
			return validationx.ValidateLogin(m.Login)
		})),
		validation.Field(&m.Password, validation.By(func(any) error {
			return validationx.ValidatePassword(m.Password, m.Login)
		})),
	)
}

type LoginRequest struct {
	User
}

func (m LoginRequest) Validate() error {
	return validation.ValidateStruct(&m,
		validation.Field(&m.Login, validation.Required),
		validation.Field(&m.Password, validation.Required),
	)
}
