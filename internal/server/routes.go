package server

import (
	"net/http"

	"github.com/dsnikitin/gophermart/internal/config"
	"github.com/dsnikitin/gophermart/internal/handler"
	"github.com/dsnikitin/gophermart/internal/middleware"

	"github.com/go-chi/chi/v5"
)

func initRouter(cfg *config.Config, h *handler.Handler) *chi.Mux {
	router := chi.NewRouter()

	router.Use(middleware.Logger)
	router.Route("/api/user", func(r chi.Router) {
		r.Post("/register", http.HandlerFunc(h.User.Register))
		r.Post("/login", http.HandlerFunc(h.User.Login))

		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(cfg.Auth))

			r.Get("/orders", http.HandlerFunc(h.Order.GetOrders))
			r.Post("/orders", http.HandlerFunc(h.Order.UploadOrder))
			r.Get("/balance", http.HandlerFunc(h.Balance.GetBalance))
			r.Post("/balance/withdraw", http.HandlerFunc(h.Balance.Withdraw))
			r.Get("/withdrawals", http.HandlerFunc(h.Balance.GetWithdrawals))
		})
	})

	return router
}
