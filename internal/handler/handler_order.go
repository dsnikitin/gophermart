package handler

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/dsnikitin/gophermart/internal/models"
	"github.com/dsnikitin/gophermart/internal/pkg/errx"
	"github.com/dsnikitin/gophermart/internal/pkg/logger"
	"github.com/pkg/errors"
)

type OrderService interface {
	UploadOrder(ctx context.Context, login, orderNumber string) error
	GetOrders(ctx context.Context, login string) ([]models.Order, error)
}

type AccrualService interface {
	NotifyOrderUploaded()
}

type OrderHandler struct {
	order   OrderService
	accrual AccrualService
}

func NewOrderHandler(order OrderService, accrual AccrualService) *OrderHandler {
	return &OrderHandler{order: order, accrual: accrual}
}

func (h *OrderHandler) UploadOrder(w http.ResponseWriter, r *http.Request) {
	login := r.Header.Get("x-user-login")

	numberBytes, err := io.ReadAll(r.Body)
	if err != nil {
		err = errors.Wrap(err, "read all")
		logger.Log.Errorw("Failed to read create order request body", "error", err.Error())
		http.Error(w, errx.ErrInternalServer.Error(), http.StatusInternalServerError)
		return
	}

	if len(numberBytes) == 0 {
		http.Error(w, "empty request body", http.StatusBadRequest)
		return
	}

	number := string(numberBytes)
	if err = h.order.UploadOrder(r.Context(), login, number); err != nil {
		err = errors.Wrap(err, "upload order")

		switch {
		case errors.Is(err, errx.ErrAlreadyAccepted):
			w.WriteHeader(http.StatusOK)
		case errors.Is(err, errx.ErrAlreadyExists):
			w.WriteHeader(http.StatusConflict)
		case errors.Is(err, errx.ErrInvalidOrderNumber):
			logger.Log.Infow("Failed to create order", "user", login, "number", number, "error", err.Error())
			http.Error(w, errx.ErrInvalidOrderNumber.Error(), http.StatusUnprocessableEntity)
		default:
			logger.Log.Errorw("Failed to create order", "user", login, "number", number, "error", err.Error())
			http.Error(w, errx.ErrInternalServer.Error(), http.StatusInternalServerError)
		}

		return
	}

	go h.accrual.NotifyOrderUploaded()

	w.WriteHeader(http.StatusAccepted)
}

func (h *OrderHandler) GetOrders(w http.ResponseWriter, r *http.Request) {
	login := r.Header.Get("x-user-login")

	orders, err := h.order.GetOrders(r.Context(), login)
	if err != nil {
		err = errors.Wrap(err, "get orders")
		logger.Log.Errorw("Failed to get orders", "user", login, "error", err.Error())
		http.Error(w, errx.ErrInternalServer.Error(), http.StatusInternalServerError)
		return
	}

	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err = json.NewEncoder(w).Encode(orders); err != nil {
		err = errors.Wrap(err, "encode")
		logger.Log.Errorw("Failed to encode user orders response", "error", err.Error())
	}
}
