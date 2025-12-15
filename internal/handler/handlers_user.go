package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/dsnikitin/gophermart/internal/models"
	"github.com/dsnikitin/gophermart/internal/pkg/auth"
	"github.com/dsnikitin/gophermart/internal/pkg/errx"
	"github.com/dsnikitin/gophermart/internal/pkg/logger"

	"github.com/pkg/errors"
)

type UserService interface {
	Register(ctx context.Context, login, password string) error
	Login(ctx context.Context, login, password string) error
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Log.Errorw("Failed to decode register request", "error", err.Error())
		http.Error(w, errx.ErrInternalServer.Error(), http.StatusInternalServerError)
		return
	}

	if err := req.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err := h.user.Register(r.Context(), req.Login, req.Password); err != nil {
		if !errors.Is(err, errx.ErrAlreadyExists) {
			logger.Log.Errorw("Failed to register user", "user", req.Login, "error", err.Error())
			http.Error(w, errx.ErrInternalServer.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusConflict)
		return
	}

	if err := h.setAuthCookie(w, req.Login); err != nil {
		logger.Log.Errorw("Failed to set aut cookie after registration", "user", req.Login, "error", err.Error())
		http.Error(w, errx.ErrInternalServer.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Log.Errorw("Failed to decode login request", "error", err.Error())
		http.Error(w, errx.ErrInternalServer.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err := req.Validate(); err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if err := h.user.Login(r.Context(), req.Login, req.Password); err != nil {
		if !errors.Is(err, errx.ErrNotFound) {
			logger.Log.Errorw("Failed to login user", "user", req.Login, "error", err.Error())
			http.Error(w, errx.ErrInternalServer.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if err := h.setAuthCookie(w, req.Login); err != nil {
		logger.Log.Errorw("Failed to set aut cookie after login", "user", req.Login, "error", err.Error())
		http.Error(w, errx.ErrInternalServer.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) setAuthCookie(w http.ResponseWriter, login string) error {
	authToken, err := auth.CreateToken(h.cfg.Auth, login)
	if err != nil {
		return errors.Wrap(err, "create token")
	}

	cookie := &http.Cookie{
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Secure:   h.cfg.Log.EnvType == logger.ProdEnv,
		Name:     h.cfg.Auth.CookieName,
		Value:    authToken,
	}

	http.SetCookie(w, cookie)

	return nil
}
