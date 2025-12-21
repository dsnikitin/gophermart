package errx

import "github.com/pkg/errors"

var (
	ErrInternalServer     = errors.New("internal server error")
	ErrAlreadyExists      = errors.New("already exists")
	ErrAlreadyAccepted    = errors.New("already accepted")
	ErrInvalidOrderNumber = errors.New("invalid order number")
	ErrInvalidPassword    = errors.New("invalid password")
	ErrNotFound           = errors.New("not found")
	ErrInsufficientFunds  = errors.New("insufficient funds")
	ErrAllWorkersBusy     = errors.New("all workers are busy")
	ErrToManyRequests     = errors.New("too many requests")
	ErrUnregisteredOrder  = errors.New("unregistered order")
)
