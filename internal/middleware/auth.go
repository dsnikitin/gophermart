package middleware

import (
	"context"
	"net/http"

	"github.com/dsnikitin/gophermart/internal/pkg/auth"
	"github.com/dsnikitin/gophermart/internal/pkg/logger"
)

func Auth(cfg *auth.Config) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(cfg.CookieName)
			if err != nil {
				logger.Log.Info("Empty auth cookie")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			claims, err := auth.ParseToken(cfg, cookie.Value)
			if err != nil {
				logger.Log.Errorw("Failed to parse auth token", "error", err)
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			if claims.Login == "" {
				logger.Log.Info("Empty login in auth token")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), "x-user-login", claims.Login)
			h.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
