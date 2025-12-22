package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/dsnikitin/gophermart/internal/config"
	"github.com/dsnikitin/gophermart/internal/models"
	"github.com/dsnikitin/gophermart/internal/pkg/auth"
	"github.com/dsnikitin/gophermart/internal/pkg/errx"
	"github.com/dsnikitin/gophermart/internal/pkg/logger"

	"github.com/pkg/errors"
)

type UserService interface {
	Register(ctx context.Context, req models.RegisterRequest) error
	Login(ctx context.Context, req models.LoginRequest) error
}

type UserHandler struct {
	cfg     *config.Config
	service UserService
}

func NewUserHandler(cfg *config.Config, service UserService) *UserHandler {
	return &UserHandler{cfg: cfg, service: service}
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		err = errors.Wrap(err, "decode")
		logger.Log.Errorw("Failed to decode register request body", "error", err.Error())
		http.Error(w, errx.ErrInternalServer.Error(), http.StatusInternalServerError)
		return
	}

	if err := req.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.service.Register(r.Context(), req); err != nil {
		err = errors.Wrap(err, "register")
		if !errors.Is(err, errx.ErrAlreadyExists) {
			logger.Log.Errorw("Failed to register user", "user", req.Login, "error", err.Error())
			http.Error(w, errx.ErrInternalServer.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusConflict)
		return
	}

	if err := h.setAuthCookie(w, req.Login); err != nil {
		err = errors.Wrap(err, "set auth cookie")
		logger.Log.Errorw("Failed to set auth cookie after registration", "user", req.Login, "error", err.Error())
		http.Error(w, errx.ErrInternalServer.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		err = errors.Wrap(err, "decode")
		logger.Log.Errorw("Failed to decode login request body", "error", err.Error())
		http.Error(w, errx.ErrInternalServer.Error(), http.StatusInternalServerError)
		return
	}

	if err := req.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.service.Login(r.Context(), req); err != nil {
		err = errors.Wrap(err, "login")
		if !errors.Is(err, errx.ErrNotFound) && !errors.Is(err, errx.ErrInvalidPassword) {
			logger.Log.Errorw("Failed to login user", "user", req.Login, "error", err.Error())
			http.Error(w, errx.ErrInternalServer.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if err := h.setAuthCookie(w, req.Login); err != nil {
		err = errors.Wrap(err, "set auth cookie")
		logger.Log.Errorw("Failed to set aut cookie after login", "user", req.Login, "error", err.Error())
		http.Error(w, errx.ErrInternalServer.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *UserHandler) setAuthCookie(w http.ResponseWriter, login string) error {
	authToken, err := auth.CreateToken(h.cfg.Auth, login)
	if err != nil {
		return errors.Wrap(err, "create token")
	}

	cookie := &http.Cookie{
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   h.cfg.Log.IsProduction,
		Name:     h.cfg.Auth.CookieName,
		Value:    authToken,
	}

	http.SetCookie(w, cookie)

	return nil
}
