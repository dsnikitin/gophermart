package handler

import (
	"context"
	"io"
	"net/http"

	"github.com/dsnikitin/gophermart/internal/models"
	"github.com/dsnikitin/gophermart/internal/pkg/errx"
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
	login, err := getLogin(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, errx.ErrInternalServer, "error", err.Error())
		return
	}

	numberBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
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
			writeError(w, http.StatusUnprocessableEntity, errx.ErrInvalidOrderNumber, "user", login, "number", number, "error", err.Error())
		default:
			writeError(w, http.StatusInternalServerError, errx.ErrInternalServer, "user", login, "number", number, "error", err.Error())
		}

		return
	}

	go h.accrual.NotifyOrderUploaded()

	w.WriteHeader(http.StatusAccepted)
}

func (h *OrderHandler) GetOrders(w http.ResponseWriter, r *http.Request) {
	login, err := getLogin(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, errx.ErrInternalServer, "error", err.Error())
		return
	}

	orders, err := h.order.GetOrders(r.Context(), login)
	if err != nil {
		err = errors.Wrap(err, "get orders")
		writeError(w, http.StatusInternalServerError, errx.ErrInternalServer, "user", login, "error", err.Error())
		return
	}

	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	writeJSON(w, http.StatusOK, orders)
}
