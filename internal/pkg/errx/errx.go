package errx

import "github.com/pkg/errors"

var (
	ErrInternalServer     = errors.New("internal server error")
	ErrAlreadyExists      = errors.New("already exists")
	ErrUserOrderExists    = errors.New("user order exists")
	ErrInvalidOrderNumber = errors.New("invalid order number")
	ErrInvalidPassword    = errors.New("invalid password")
	ErrNotFound           = errors.New("not found")
)
