package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/dsnikitin/gophermart/internal/models"
	"github.com/dsnikitin/gophermart/internal/pkg/errx"
	"github.com/pkg/errors"
)

type BalanceService interface {
	GetBalance(ctx context.Context, login string) (models.Balance, error)
	Withdraw(ctx context.Context, login string, withdrawal models.WithdrawRequest) error
	GetWithdrawals(ctx context.Context, login string) ([]models.Withdrawal, error)
}

type BalanceHandler struct {
	service BalanceService
}

func NewBalanceHandler(service BalanceService) *BalanceHandler {
	return &BalanceHandler{service: service}
}

func (h *BalanceHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	login, err := getLogin(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, errx.ErrInternalServer, "error", err.Error())
		return
	}

	balance, err := h.service.GetBalance(r.Context(), login)
	if err != nil {
		err := errors.Wrap(err, "get balance")
		writeError(w, http.StatusInternalServerError, errx.ErrInternalServer, "user", login, "error", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, balance)
}

func (h *BalanceHandler) Withdraw(w http.ResponseWriter, r *http.Request) {
	login, err := getLogin(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, errx.ErrInternalServer, "error", err.Error())
		return
	}

	var req models.WithdrawRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := req.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.service.Withdraw(r.Context(), login, req); err != nil {
		err := errors.Wrap(err, "withdraw")

		switch {
		case errors.Is(err, errx.ErrInsufficientFunds):
			w.WriteHeader(http.StatusPaymentRequired)
		case errors.Is(err, errx.ErrInvalidOrderNumber):
			writeError(w, http.StatusUnprocessableEntity, errx.ErrInvalidOrderNumber, "user", login, "request", req, "error", err.Error())
		default:
			writeError(w, http.StatusInternalServerError, errx.ErrInternalServer, "user", login, "request", req, "error", err.Error())
		}

		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *BalanceHandler) GetWithdrawals(w http.ResponseWriter, r *http.Request) {
	login, err := getLogin(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, errx.ErrInternalServer, "error", err.Error())
		return
	}

	withdrawals, err := h.service.GetWithdrawals(r.Context(), login)
	if err != nil {
		err := errors.Wrap(err, "get withdrawals")
		writeError(w, http.StatusInternalServerError, errx.ErrInternalServer, "user", login, "error", err.Error())
	}

	if len(withdrawals) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	writeJSON(w, http.StatusOK, withdrawals)
}
