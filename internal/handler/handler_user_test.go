package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dsnikitin/gophermart/internal/config"
	"github.com/dsnikitin/gophermart/internal/handler"
	"github.com/dsnikitin/gophermart/internal/models"
	"github.com/dsnikitin/gophermart/internal/pkg/auth"
	"github.com/dsnikitin/gophermart/internal/pkg/errx"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) Register(ctx context.Context, req models.RegisterRequest) error {
	args := m.Called(req)
	return args.Error(0)
}

func (m *MockUserService) Login(ctx context.Context, req models.LoginRequest) error {
	args := m.Called(req)
	return args.Error(0)
}

func TestUserHandler_Register(t *testing.T) {
	type headers struct {
		contentType string
		setCookie   string
	}

	type want struct {
		code    int
		headers headers
		errMsg  string
	}

	cfg := &config.Config{
		Auth: &auth.Config{
			SigningKey: "testkey",
			TokenExp:   time.Hour,
			CookieName: "auth_token",
		},
	}

	s := new(MockUserService)
	h := handler.New(cfg, &handler.Service{User: s})

	r := chi.NewRouter()
	r.Post("/api/user/register", h.User.Register)

	regReq := models.RegisterRequest{User: models.User{Login: "user1", Password: "Qwerty123+"}}

	tests := []struct {
		name             string
		method           string
		req              models.RegisterRequest
		servicesMockCall func()
		want             want
	}{
		{
			name:   "positive",
			method: http.MethodPost,
			req:    regReq,
			servicesMockCall: func() {
				s.On("Register", regReq).Return(nil).Once()
			},
			want: want{
				code:    http.StatusOK,
				headers: headers{setCookie: cfg.Auth.CookieName + "="},
			},
		},
		{
			name:             "wrong method",
			method:           http.MethodGet,
			req:              regReq,
			servicesMockCall: func() {},
			want: want{
				code: http.StatusMethodNotAllowed,
			},
		},
		{
			name:             "empty login",
			method:           http.MethodPost,
			req:              models.RegisterRequest{User: models.User{Login: "", Password: "Qwerty123+"}},
			servicesMockCall: func() {},
			want: want{
				code:    http.StatusBadRequest,
				headers: headers{contentType: "text/plain; charset=utf-8"},
				errMsg:  "login required",
			},
		},
		{
			name:             "empty password",
			method:           http.MethodPost,
			req:              models.RegisterRequest{User: models.User{Login: "user1", Password: ""}},
			servicesMockCall: func() {},
			want: want{
				code:    http.StatusBadRequest,
				headers: headers{contentType: "text/plain; charset=utf-8"},
				errMsg:  "password required",
			},
		},
		{
			name:   "already exist",
			method: http.MethodPost,
			req:    regReq,
			servicesMockCall: func() {
				s.On("Register", regReq).Return(errx.ErrAlreadyExists).Once()
			},
			want: want{
				code: http.StatusConflict,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.servicesMockCall()

			body, err := json.Marshal(test.req)
			require.NoError(t, err)

			req := httptest.NewRequest(test.method, "/api/user/register", bytes.NewBuffer(body))
			req.Header.Add("Content-Type", "application/json")

			recorder := httptest.NewRecorder()
			r.ServeHTTP(recorder, req)

			res := recorder.Result()
			defer res.Body.Close()

			resBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)

			assert.Equal(t, test.want.code, res.StatusCode)

			switch res.StatusCode {
			case http.StatusOK:
				cookie := res.Header.Get("Set-Cookie")
				assert.NotEmpty(t, cookie, "auth cookie should be not empty")
				assert.Contains(t, cookie, cfg.Auth.CookieName+"=")
			case http.StatusBadRequest:
				assert.Equal(t, test.want.headers.contentType, res.Header.Get("Content-Type"))
				assert.NotEmpty(t, resBody, "response body should not be empty")
			case http.StatusConflict:
				assert.Empty(t, resBody, "response body should be empty")
			}
		})
	}
}

func TestUserHandler_Login(t *testing.T) {
	type headers struct {
		contentType string
		setCookie   string
	}

	type want struct {
		code    int
		headers headers
		resBody string
	}

	cfg := &config.Config{
		Auth: &auth.Config{
			SigningKey: "testkey",
			TokenExp:   time.Hour,
			CookieName: "auth_token",
		},
	}

	s := new(MockUserService)
	h := handler.New(cfg, &handler.Service{User: s})

	r := chi.NewRouter()
	r.Post("/api/user/login", h.User.Login)

	loginReq := models.LoginRequest{User: models.User{Login: "user1", Password: "Qwerty123+"}}

	tests := []struct {
		name             string
		method           string
		req              models.LoginRequest
		servicesMockCall func()
		want             want
	}{
		{
			name:   "positive",
			method: http.MethodPost,
			req:    loginReq,
			servicesMockCall: func() {
				s.On("Login", loginReq).Return(nil).Once()
			},
			want: want{
				code:    http.StatusOK,
				headers: headers{setCookie: cfg.Auth.CookieName + "="},
			},
		},
		{
			name:             "wrong method",
			method:           http.MethodGet,
			req:              loginReq,
			servicesMockCall: func() {},
			want: want{
				code: http.StatusMethodNotAllowed,
			},
		},
		{
			name:             "empty login",
			method:           http.MethodPost,
			req:              models.LoginRequest{User: models.User{Login: "", Password: "Qwerty123+"}},
			servicesMockCall: func() {},
			want: want{
				code:    http.StatusBadRequest,
				headers: headers{contentType: "text/plain; charset=utf-8"},
				resBody: "login required",
			},
		},
		{
			name:             "empty password",
			method:           http.MethodPost,
			req:              models.LoginRequest{User: models.User{Login: "user1", Password: ""}},
			servicesMockCall: func() {},
			want: want{
				code:    http.StatusBadRequest,
				headers: headers{contentType: "text/plain; charset=utf-8"},
				resBody: "password required",
			},
		},
		{
			name:   "wrong login",
			method: http.MethodPost,
			req:    loginReq,
			servicesMockCall: func() {
				s.On("Login", loginReq).Return(errx.ErrNotFound).Once()
			},
			want: want{
				code: http.StatusUnauthorized,
			},
		},
		{
			name:   "wrong password",
			method: http.MethodPost,
			req:    loginReq,
			servicesMockCall: func() {
				s.On("Login", loginReq).Return(errx.ErrInvalidPassword).Once()
			},
			want: want{
				code: http.StatusUnauthorized,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.servicesMockCall()

			body, err := json.Marshal(test.req)
			require.NoError(t, err)

			req := httptest.NewRequest(test.method, "/api/user/login", bytes.NewBuffer(body))
			req.Header.Add("Content-Type", "application/json")

			recorder := httptest.NewRecorder()
			r.ServeHTTP(recorder, req)

			res := recorder.Result()
			defer res.Body.Close()

			resBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)

			assert.Equal(t, test.want.code, res.StatusCode)
			switch res.StatusCode {
			case http.StatusOK:
				cookie := res.Header.Get("Set-Cookie")
				assert.NotEmpty(t, cookie, "auth cookie should be not empty")
				assert.Contains(t, cookie, cfg.Auth.CookieName+"=")
			case http.StatusBadRequest:
				assert.Equal(t, test.want.headers.contentType, res.Header.Get("Content-Type"))
				assert.NotEmpty(t, resBody, "response body should not be empty")
			case http.StatusUnauthorized:
				assert.Empty(t, resBody, "response body should be empty")
			}
		})
	}
}
