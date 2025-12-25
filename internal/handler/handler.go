package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/dsnikitin/gophermart/internal/config"
	"github.com/dsnikitin/gophermart/internal/models"
	"github.com/dsnikitin/gophermart/internal/pkg/logger"
	"github.com/pkg/errors"
)

type Handler struct {
	User    *UserHandler
	Order   *OrderHandler
	Balance *BalanceHandler
}

type Service struct {
	User    UserService
	Order   OrderService
	Balance BalanceService
	Accrual AccrualService
}

func New(cfg *config.Config, s *Service) *Handler {
	return &Handler{
		User:    NewUserHandler(cfg, s.User),
		Order:   NewOrderHandler(s.Order, s.Accrual),
		Balance: NewBalanceHandler(s.Balance),
	}
}

func getLogin(ctx context.Context) (string, error) {
	ctxValue := ctx.Value(models.LoginKey{})
	if ctxValue == nil {
		return "", errors.New("login not found in context")
	}

	login, ok := ctxValue.(string)
	if !ok || login == "" {
		return "", errors.Errorf("empty login value or unexpected type: value=%v, type=%T", ctxValue, ctxValue)
	}

	return login, nil
}

func writeError(w http.ResponseWriter, code int, err error, keysAndValues ...any) {
	logger.Log.Errorw("Failed to handle request", keysAndValues...)
	http.Error(w, err.Error(), code)
}

func writeJSON(w http.ResponseWriter, code int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		err = errors.Wrap(err, "encode")
		logger.Log.Errorw("Failed to encode json response", "error", err.Error())
	}
}
