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
	Upload(ctx context.Context, login, orderNumber string) error
	GetByUser(ctx context.Context, login string) ([]models.Order, error)
}

func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	login := r.Header.Get("x-user-login")

	orderNumber, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Log.Errorw("Failed to read create order request body", "error", err.Error())
		http.Error(w, errx.ErrInternalServer.Error(), http.StatusInternalServerError)
		return
	}

	if len(orderNumber) == 0 {
		http.Error(w, "empty request body", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain")

	if err := h.order.Upload(r.Context(), login, string(orderNumber)); err != nil {
		err := errors.Wrap(err, "order upload")

		switch {
		case errors.Is(err, errx.ErrUserOrderExists):
			w.WriteHeader(http.StatusOK)
		case errors.Is(err, errx.ErrAlreadyExists):
			w.WriteHeader(http.StatusConflict)
		case errors.Is(err, errx.ErrInvalidOrderNumber):
			logger.Log.Infow("Failed to create order", "user", login, "number", string(orderNumber), "error", err.Error())
			http.Error(w, errx.ErrInvalidOrderNumber.Error(), http.StatusUnprocessableEntity)
		default:
			logger.Log.Errorw("Failed to create order", "user", login, "number", string(orderNumber), "error", err.Error())
			http.Error(w, errx.ErrInternalServer.Error(), http.StatusInternalServerError)
		}

		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (h *Handler) GetOrders(w http.ResponseWriter, r *http.Request) {
	login := r.Header.Get("x-user-login")

	orders, err := h.order.GetByUser(r.Context(), login)
	if err != nil {
		logger.Log.Errorw("Failed to get user orders", "user", login, "error", err.Error())
		http.Error(w, errx.ErrInternalServer.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(orders); err != nil {
		logger.Log.Errorw("Failed to encode user orders response", "error", err)
		return
	}
}
