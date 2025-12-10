package server

import (
	"net/http"

	"github.com/dsnikitin/gophermart/internal/config"
	"github.com/dsnikitin/gophermart/internal/pkg/middleware"

	"github.com/go-chi/chi/v5"
)

type Handler interface {
	Register(w http.ResponseWriter, r *http.Request)
	Login(w http.ResponseWriter, r *http.Request)
	CreateOrder(w http.ResponseWriter, r *http.Request)
	GetOrders(w http.ResponseWriter, r *http.Request)
	GetBalance(w http.ResponseWriter, r *http.Request)
	Withdraw(w http.ResponseWriter, r *http.Request)
	GetWithdrawals(w http.ResponseWriter, r *http.Request)
}

func InitChiMux(cfg *config.Config, h Handler) *chi.Mux {
	router := chi.NewRouter()
	router.Use(middleware.Logging)

	router.Route("/api/user", func(r chi.Router) {
		r.Post("/register", http.HandlerFunc(h.Register))
		r.Post("/login", http.HandlerFunc(h.Login))

		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(cfg.Auth))

			r.Get("/orders", http.HandlerFunc(h.GetOrders))
			r.Post("/orders", http.HandlerFunc(h.CreateOrder))
			r.Get("/balance", http.HandlerFunc(h.GetBalance))
			r.Post("/balance/withdraw", http.HandlerFunc(h.Withdraw))
			r.Get("/withdrawals", http.HandlerFunc(h.GetWithdrawals))
		})
	})

	return router
}
