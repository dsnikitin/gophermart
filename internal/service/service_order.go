package service

import (
	"context"

	"github.com/dsnikitin/gophermart/internal/models"
	"github.com/dsnikitin/gophermart/internal/pkg/accrualer"
	"github.com/dsnikitin/gophermart/internal/pkg/errx"
	"github.com/pkg/errors"
)

type OrderRepository interface {
	UploadOrder(ctx context.Context, login, number string) error
	GetOrder(ctx context.Context, number string) (models.Order, error)
	GetOrders(ctx context.Context, login string) ([]models.Order, error)
}

type OrderService struct {
	r         OrderRepository
	accrualer *accrualer.Accrualer
}

func NewOrder(r OrderRepository) *OrderService {
	return &OrderService{r: r}
}

func (s *OrderService) UploadOrder(ctx context.Context, login, number string) error {
	if err := checkNumber(number); err != nil {
		return errors.Wrap(err, "check number")
	}

	if err := s.r.UploadOrder(ctx, login, number); err != nil {
		if !errors.Is(err, errx.ErrAlreadyExists) {
			return errors.Wrap(err, "upload order")
		}

		order, err := s.r.GetOrder(ctx, number)
		if err != nil {
			return errors.Wrap(err, "get order")
		}

		if order.UserLogin == login {
			return errx.ErrAlreadyAccepted
		}

		return errx.ErrAlreadyExists
	}

	return nil
}

func (s *OrderService) GetOrders(ctx context.Context, login string) ([]models.Order, error) {
	return s.r.GetOrders(ctx, login)
}

func checkNumber(number string) error {
	if len(number) < 2 {
		return errors.Wrap(errx.ErrInvalidOrderNumber, "number must be 2 characters long")
	}

	onlyZeros := true
	for i := 0; i < len(number); i++ {
		ch := number[i]

		if ch < '0' || ch > '9' {
			return errors.Wrap(errx.ErrInvalidOrderNumber, "number must contain only digits 0-9")
		}

		if onlyZeros && ch != '0' {
			onlyZeros = false
		}
	}

	if onlyZeros {
		return errors.Wrap(errx.ErrInvalidOrderNumber, "number cannot contain only zeros")
	}

	err := checkLuhn(number)
	return errors.Wrap(err, "check luhn")
}

func checkLuhn(number string) error {
	sum := 0
	shouldDouble := false
	for i := len(number) - 1; i >= 0; i-- {
		digit := int(number[i] - '0')
		if shouldDouble {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}

		sum += digit
		shouldDouble = !shouldDouble
	}

	if sum%10 != 0 {
		return errx.ErrInvalidOrderNumber
	}

	return nil
}
