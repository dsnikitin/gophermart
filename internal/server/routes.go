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
	router.Use(middleware.GzipCompress)

	router.Route("/api/user", func(public chi.Router) {
		public.Post("/register", http.HandlerFunc(h.User.Register))
		public.Post("/login", http.HandlerFunc(h.User.Login))

		public.Group(func(protected chi.Router) {
			protected.Use(middleware.Auth(cfg.Auth))

			protected.Get("/orders", http.HandlerFunc(h.Order.GetOrders))
			protected.Post("/orders", http.HandlerFunc(h.Order.UploadOrder))
			protected.Get("/balance", http.HandlerFunc(h.Balance.GetBalance))
			protected.Post("/balance/withdraw", http.HandlerFunc(h.Balance.Withdraw))
			protected.Get("/withdrawals", http.HandlerFunc(h.Balance.GetWithdrawals))
		})
	})

	return router
}
