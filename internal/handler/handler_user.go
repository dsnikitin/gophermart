package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/dsnikitin/gophermart/internal/config"
	"github.com/dsnikitin/gophermart/internal/models"
	"github.com/dsnikitin/gophermart/internal/pkg/auth"
	"github.com/dsnikitin/gophermart/internal/pkg/errx"

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
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := req.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.service.Register(r.Context(), req); err != nil {
		if !errors.Is(err, errx.ErrAlreadyExists) {
			err = errors.Wrap(err, "register")
			writeError(w, http.StatusInternalServerError, errx.ErrInternalServer, "user", req.Login, "error", err.Error())
			return
		}

		w.WriteHeader(http.StatusConflict)
		return
	}

	if err := h.setAuthCookie(w, req.Login); err != nil {
		err = errors.Wrap(err, "set auth cookie")
		writeError(w, http.StatusInternalServerError, errx.ErrInternalServer, "user", req.Login, "error", err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := req.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.service.Login(r.Context(), req); err != nil {
		if !errors.Is(err, errx.ErrNotFound) && !errors.Is(err, errx.ErrInvalidPassword) {
			err = errors.Wrap(err, "login")
			writeError(w, http.StatusInternalServerError, errx.ErrInternalServer, "user", req.Login, "error", err.Error())
			return
		}

		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if err := h.setAuthCookie(w, req.Login); err != nil {
		err = errors.Wrap(err, "set auth cookie")
		writeError(w, http.StatusInternalServerError, errx.ErrInternalServer, "user", req.Login, "error", err.Error())
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
