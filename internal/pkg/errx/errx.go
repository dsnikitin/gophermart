package errx

import "github.com/pkg/errors"

var (
	ErrInternalServer  = errors.New("internal server error")
	ErrAlreadyExists   = errors.New("already exists")
	ErrInvalidPassword = errors.New("invalid password")
	ErrNotFound        = errors.New("not found")
)
