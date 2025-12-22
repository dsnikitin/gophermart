package app

import (
	"context"
	"net/http"
	"time"

	"github.com/dsnikitin/gophermart/internal/config"
	"github.com/dsnikitin/gophermart/internal/handler"
	"github.com/dsnikitin/gophermart/internal/pkg/adapter"
	"github.com/dsnikitin/gophermart/internal/pkg/logger"
	"github.com/dsnikitin/gophermart/internal/repository"
	"github.com/dsnikitin/gophermart/internal/server"
	"github.com/dsnikitin/gophermart/internal/service"
	"github.com/jackc/pgx/v5/pgxpool"
)

type App struct {
	cfg     *config.Config
	srv     *server.Server
	service *service.Service
}

func New(cfg *config.Config, db *pgxpool.Pool) *App {
	repo := repository.New(db)

	services := service.New(cfg, &service.Dependencies{
		User:      repo.User,
		Order:     repo.Order,
		Balance:   repo.Balance,
		BalanceTx: adapter.NewBalanceTxAdapter(repo.Balance),
		Accrual:   repo.Accrual,
		AccrualTx: adapter.NewAccrualTxAdapter(repo.Accrual),
	})

	handlers := handler.New(cfg, &handler.Service{
		User:    services.User,
		Order:   services.Order,
		Balance: services.Balance,
		Accrual: services.Accrual,
	})

	return &App{
		cfg:     cfg,
		srv:     server.New(cfg, handlers),
		service: services,
	}
}

func (a *App) Run() {
	logger.Log.Infow("Starting server", "address", a.cfg.ServerAddr)
	if err := a.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Log.Fatalw("Failed to start server", "error", err)
	}
}

func (a *App) Shutdown() {
	a.service.Accrual.Stop()

	logger.Log.Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	if err := a.srv.Shutdown(ctx); err != nil {
		logger.Log.Errorw("Failed to shutdown server", "error", err)
	} else {
		logger.Log.Info("Server shutdown gracefully")
	}
}
