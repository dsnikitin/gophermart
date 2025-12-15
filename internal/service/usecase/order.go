package usecase

import (
	"context"
	"time"

	"github.com/dsnikitin/gophermart/internal/models"
	"github.com/dsnikitin/gophermart/internal/pkg/consts/status"
	"github.com/dsnikitin/gophermart/internal/pkg/errx"
	"github.com/pkg/errors"
)

type OrderRepository interface {
	UploadOrder(ctx context.Context, newOrder models.Order) (models.Order, error)
	GetUserOrders(ctx context.Context, login string) ([]models.Order, error)
}

type Order struct {
	r OrderRepository
}

func NewOrder(r OrderRepository) *Order {
	return &Order{r: r}
}

func (s *Order) Upload(ctx context.Context, login, number string) error {
	if err := checkNumber(number); err != nil {
		return errors.Wrap(err, "check number")
	}

	newOrder := models.Order{
		Number:     number,
		Status:     status.New,
		UploadedAt: time.Now().UTC(),
		UserLogin:  login,
	}

	order, err := s.r.UploadOrder(ctx, newOrder)
	if err != nil {
		return errors.Wrap(err, "upload order")
	}

	if order.UserLogin != login {
		return errx.ErrAlreadyExists
	}

	if order.UploadedAt != newOrder.UploadedAt {
		return errx.ErrUserOrderExists
	}

	return nil
}

func (s *Order) GetByUser(ctx context.Context, login string) ([]models.Order, error) {
	return s.r.GetUserOrders(ctx, login)
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
