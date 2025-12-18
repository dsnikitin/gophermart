package service

import (
	"context"
	"time"

	"github.com/dsnikitin/gophermart/internal/models"
	"github.com/dsnikitin/gophermart/internal/pkg/consts/status"
	"github.com/dsnikitin/gophermart/internal/pkg/errx"
	"github.com/pkg/errors"
)

type OrderRepository interface {
	UploadOrder(ctx context.Context, newOrder models.OrderDB) (*models.OrderDB, error)
	GetOrders(ctx context.Context, login string) ([]*models.OrderDB, error)
}

type OrderService struct {
	r OrderRepository
}

func NewOrder(r OrderRepository) *OrderService {
	return &OrderService{r: r}
}

func (s *OrderService) UploadOrder(ctx context.Context, login, number string) error {
	if err := checkNumber(number); err != nil {
		return errors.Wrap(err, "check number")
	}

	newOrder := models.OrderDB{
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

func (s *OrderService) GetOrders(ctx context.Context, login string) ([]models.OrderResponse, error) {
	orders, err := s.r.GetOrders(ctx, login)
	if err != nil {
		return nil, errors.Wrap(err, "get orders")
	}

	res := make([]models.OrderResponse, 0, len(orders))
	for _, o := range orders {
		res = append(res, o.ToOrderResponse())
	}

	return res, nil
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
