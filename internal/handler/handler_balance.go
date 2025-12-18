package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/dsnikitin/gophermart/internal/models"
	"github.com/dsnikitin/gophermart/internal/pkg/errx"
	"github.com/dsnikitin/gophermart/internal/pkg/logger"
	"github.com/pkg/errors"
)

type BalanceService interface {
	GetBalance(ctx context.Context, login string) (models.BalanceResponse, error)
	Withdraw(ctx context.Context, login string, withdrawal models.WithdrawRequest) error
	GetWithdrawals(ctx context.Context, login string) ([]models.WithdrawalResponse, error)
}

type BalanceHandler struct {
	service BalanceService
}

func NewBalance(service BalanceService) *BalanceHandler {
	return &BalanceHandler{service: service}
}

func (h *BalanceHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	login := r.Header.Get("x-user-login")

	balance, err := h.service.GetBalance(r.Context(), login)
	if err != nil {
		err = errors.Wrap(err, "get balance")
		logger.Log.Errorw("Failed to get balance", "user", login, "error", err.Error())
		http.Error(w, errx.ErrInternalServer.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(balance); err != nil {
		err = errors.Wrap(err, "encode")
		logger.Log.Errorw("Failed to encode get balance response", "error", err.Error())
	}
}

func (h *BalanceHandler) Withdraw(w http.ResponseWriter, r *http.Request) {
	login := r.Header.Get("x-user-login")

	var req models.WithdrawRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		err = errors.Wrap(err, "decode")
		logger.Log.Errorw("Failed to read withdraw request body", "error", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := req.Validate(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := h.service.Withdraw(r.Context(), login, req); err != nil {
		err := errors.Wrap(err, "withdraw")

		switch {
		case errors.Is(err, errx.ErrInsufficientFunds):
			w.WriteHeader(http.StatusPaymentRequired)
		case errors.Is(err, errx.ErrInvalidOrderNumber):
			logger.Log.Infow("Failed to withdraw", "user", login, "request", req, "error", err.Error())
			http.Error(w, errx.ErrInvalidOrderNumber.Error(), http.StatusUnprocessableEntity)
		default:
			logger.Log.Errorw("Failed to withdraw", "user", login, "request", req, "error", err.Error())
			http.Error(w, errx.ErrInternalServer.Error(), http.StatusInternalServerError)
		}

		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *BalanceHandler) GetWithdrawals(w http.ResponseWriter, r *http.Request) {
	login := r.Header.Get("x-user-login")

	withdrawals, err := h.service.GetWithdrawals(r.Context(), login)
	if err != nil {
		err := errors.Wrap(err, "get withdrawals")
		logger.Log.Errorw("Failed to get withdrawals", "user", login, "error", err.Error())
		http.Error(w, errx.ErrInternalServer.Error(), http.StatusInternalServerError)
	}

	if len(withdrawals) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(withdrawals); err != nil {
		err = errors.Wrap(err, "encode")
		logger.Log.Errorw("Failed to encode get withdrawals response", "error", err.Error())
	}
}
