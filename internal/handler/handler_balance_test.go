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
	"github.com/dsnikitin/gophermart/internal/pkg/errx"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockBalanceService struct {
	mock.Mock
}

func (m *MockBalanceService) GetBalance(ctx context.Context, login string) (models.Balance, error) {
	args := m.Called(login)
	return args.Get(0).(models.Balance), args.Error(1)
}

func (m *MockBalanceService) Withdraw(ctx context.Context, login string, req models.WithdrawRequest) error {
	args := m.Called(login, req)
	return args.Error(0)
}

func (m *MockBalanceService) GetWithdrawals(ctx context.Context, login string) ([]models.Withdrawal, error) {
	args := m.Called(login)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Withdrawal), args.Error(1)
}

func TestBalanceHandler_GetBalance(t *testing.T) {
	type headers struct {
		contentType string
	}

	type want struct {
		code    int
		headers headers
		resp    models.Balance
	}

	s := new(MockBalanceService)
	h := handler.New(&config.Config{}, &handler.Service{Balance: s})

	r := chi.NewRouter()
	r.Get("/api/user/balance", h.Balance.GetBalance)

	userLogin := "user1"
	balance := models.Balance{
		Current:   100.5,
		Withdrawn: 77.53,
	}

	tests := []struct {
		name             string
		method           string
		userLogin        string
		servicesMockCall func()
		want             want
	}{
		{
			name:      "positive",
			method:    http.MethodGet,
			userLogin: userLogin,
			servicesMockCall: func() {
				s.On("GetBalance", userLogin).Return(balance, nil).Once()
			},
			want: want{
				code:    http.StatusOK,
				headers: headers{contentType: "application/json"},
				resp:    balance,
			},
		},
		{
			name:             "wrong method",
			method:           http.MethodPost,
			userLogin:        userLogin,
			servicesMockCall: func() {},
			want: want{
				code: http.StatusMethodNotAllowed,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.servicesMockCall()

			req := httptest.NewRequest(test.method, "/api/user/balance", nil)
			if test.userLogin != "" {
				ctx := context.WithValue(req.Context(), models.LoginKey{}, test.userLogin)
				req = req.WithContext(ctx)
			}
			recorder := httptest.NewRecorder()

			r.ServeHTTP(recorder, req)
			res := recorder.Result()
			defer res.Body.Close()

			assert.Equal(t, test.want.code, res.StatusCode)
			switch res.StatusCode {
			case http.StatusOK:
				var resp models.Balance
				require.NoError(t, json.NewDecoder(res.Body).Decode(&resp))
				assert.Equal(t, test.want.resp, resp)
			}
		})
	}
}

func TestBalanceHandler_Withdraw(t *testing.T) {
	type headers struct {
		contentType string
	}

	type want struct {
		code    int
		headers headers
		errMsg  string
	}

	s := new(MockBalanceService)
	h := handler.New(&config.Config{}, &handler.Service{Balance: s})

	r := chi.NewRouter()
	r.Post("/api/user/balance/withdraw", h.Balance.Withdraw)

	userLogin := "user1"
	withdrawReq := models.WithdrawRequest{Order: "12345678903", Sum: 20.20}

	tests := []struct {
		name             string
		method           string
		userLogin        string
		req              models.WithdrawRequest
		servicesMockCall func()
		want             want
	}{
		{
			name:      "positive",
			method:    http.MethodPost,
			userLogin: userLogin,
			req:       withdrawReq,
			servicesMockCall: func() {
				s.On("Withdraw", userLogin, withdrawReq).Return(nil).Once()
			},
			want: want{
				code: http.StatusOK,
			},
		},
		{
			name:             "wrong method",
			method:           http.MethodGet,
			userLogin:        userLogin,
			req:              withdrawReq,
			servicesMockCall: func() {},
			want: want{
				code: http.StatusMethodNotAllowed,
			},
		},
		{
			name:             "empty order number",
			method:           http.MethodPost,
			userLogin:        userLogin,
			req:              models.WithdrawRequest{Sum: 100},
			servicesMockCall: func() {},
			want: want{
				code:    http.StatusBadRequest,
				headers: headers{contentType: "text/plain; charset=utf-8"},
				errMsg:  "order number required\n",
			},
		},
		{
			name:             "empty withdrawal sum",
			method:           http.MethodPost,
			userLogin:        userLogin,
			req:              models.WithdrawRequest{Order: "12345678903"},
			servicesMockCall: func() {},
			want: want{
				code:    http.StatusBadRequest,
				headers: headers{contentType: "text/plain; charset=utf-8"},
				errMsg:  "withdrawal sum required\n",
			},
		},
		{
			name:             "negative withdrawal sum",
			method:           http.MethodPost,
			userLogin:        userLogin,
			req:              models.WithdrawRequest{Order: "12345678903", Sum: -100},
			servicesMockCall: func() {},
			want: want{
				code:    http.StatusBadRequest,
				headers: headers{contentType: "text/plain; charset=utf-8"},
				errMsg:  "withdrawal sum should be positive\n",
			},
		},
		{
			name:      "invalid order number - too short",
			method:    http.MethodPost,
			userLogin: userLogin,
			req:       models.WithdrawRequest{Order: "1", Sum: 100},
			servicesMockCall: func() {
				s.On("Withdraw", userLogin, models.WithdrawRequest{Order: "1", Sum: 100}).Return(errx.ErrInvalidOrderNumber).Once()
			},
			want: want{
				code:    http.StatusUnprocessableEntity,
				headers: headers{contentType: "text/plain; charset=utf-8"},
				errMsg:  errx.ErrInvalidOrderNumber.Error() + "\n",
			},
		},
		{
			name:      "invalid order number - has non-digits",
			method:    http.MethodPost,
			userLogin: userLogin,
			req:       models.WithdrawRequest{Order: "123a", Sum: 100},
			servicesMockCall: func() {
				s.On("Withdraw", userLogin, models.WithdrawRequest{Order: "123a", Sum: 100}).Return(errx.ErrInvalidOrderNumber).Once()
			},
			want: want{
				code:    http.StatusUnprocessableEntity,
				headers: headers{contentType: "text/plain; charset=utf-8"},
				errMsg:  errx.ErrInvalidOrderNumber.Error() + "\n",
			},
		},
		{
			name:      "invalid order number - all zeros",
			method:    http.MethodPost,
			userLogin: userLogin,
			req:       models.WithdrawRequest{Order: "000", Sum: 100},
			servicesMockCall: func() {
				s.On("Withdraw", userLogin, models.WithdrawRequest{Order: "000", Sum: 100}).Return(errx.ErrInvalidOrderNumber).Once()
			},
			want: want{
				code:    http.StatusUnprocessableEntity,
				headers: headers{contentType: "text/plain; charset=utf-8"},
				errMsg:  errx.ErrInvalidOrderNumber.Error() + "\n",
			},
		},
		{
			name:      "invalid order number - luhn failed",
			method:    http.MethodPost,
			userLogin: userLogin,
			req:       models.WithdrawRequest{Order: "123456789031", Sum: 100},
			servicesMockCall: func() {
				s.On("Withdraw", userLogin, models.WithdrawRequest{Order: "123456789031", Sum: 100}).Return(errx.ErrInvalidOrderNumber).Once()
			},
			want: want{
				code:    http.StatusUnprocessableEntity,
				headers: headers{contentType: "text/plain; charset=utf-8"},
				errMsg:  errx.ErrInvalidOrderNumber.Error() + "\n",
			},
		},
		{
			name:      "insufficient funds",
			method:    http.MethodPost,
			userLogin: userLogin,
			req:       withdrawReq,
			servicesMockCall: func() {
				s.On("Withdraw", userLogin, withdrawReq).Return(errx.ErrInsufficientFunds).Once()
			},
			want: want{
				code: http.StatusPaymentRequired,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.servicesMockCall()

			body, err := json.Marshal(test.req)
			require.NoError(t, err)

			req := httptest.NewRequest(test.method, "/api/user/balance/withdraw", bytes.NewBuffer(body))
			req.Header.Add("Content-Type", "application/json")
			if test.userLogin != "" {
				ctx := context.WithValue(req.Context(), models.LoginKey{}, test.userLogin)
				req = req.WithContext(ctx)
			}

			recorder := httptest.NewRecorder()
			r.ServeHTTP(recorder, req)

			res := recorder.Result()
			defer res.Body.Close()

			resBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)

			assert.Equal(t, test.want.code, res.StatusCode)
			switch res.StatusCode {
			case http.StatusOK, http.StatusPaymentRequired:
				assert.Empty(t, resBody, "response body should be empty")
			case http.StatusBadRequest:
				assert.Equal(t, test.want.headers.contentType, res.Header.Get("Content-Type"))
				assert.NotEmpty(t, resBody, "response body should not be empty")
			case http.StatusUnprocessableEntity:
				assert.Equal(t, test.want.headers.contentType, res.Header.Get("Content-Type"))
				assert.Equal(t, test.want.errMsg, string(resBody))
			}
		})
	}
}

func TestBalanceHandler_GetWithdrawals(t *testing.T) {
	type WithdrawalResp struct {
		OrderNumber string  `json:"order"`
		Sum         float64 `json:"sum"`
		ProcessedAt string  `json:"processed_at"`
	}

	userLogin := "user1"
	now := time.Now()
	minuteAgo := now.Add(time.Minute)
	withdrawalsResp := []WithdrawalResp{
		{
			OrderNumber: "012345678903",
			Sum:         12.3,
			ProcessedAt: minuteAgo.Truncate(time.Second).Format(time.RFC3339),
		},
		{
			OrderNumber: "12345678903",
			Sum:         2.75,
			ProcessedAt: now.Truncate(time.Second).Format(time.RFC3339),
		},
	}

	type headers struct {
		contentType string
	}

	type want struct {
		code    int
		headers headers
		resp    []WithdrawalResp
	}

	s := new(MockBalanceService)
	h := handler.New(&config.Config{}, &handler.Service{Balance: s})

	r := chi.NewRouter()
	r.Get("/api/user/withdrawals", h.Balance.GetWithdrawals)

	tests := []struct {
		name             string
		method           string
		userLogin        string
		servicesMockCall func()
		want             want
	}{
		{
			name:      "positive with withdrawals",
			method:    http.MethodGet,
			userLogin: userLogin,
			servicesMockCall: func() {
				withdrawals := []models.Withdrawal{
					{
						OrderNumber: "012345678903",
						Sum:         12.3,
						ProcessedAt: minuteAgo,
					}, {
						OrderNumber: "12345678903",
						Sum:         2.75,
						ProcessedAt: now,
					},
				}
				s.On("GetWithdrawals", userLogin).Return(withdrawals, nil).Once()
			},
			want: want{
				code:    http.StatusOK,
				headers: headers{contentType: "application/json"},
				resp:    withdrawalsResp,
			},
		},
		{
			name:      "positive no content",
			method:    http.MethodGet,
			userLogin: userLogin,
			servicesMockCall: func() {
				s.On("GetWithdrawals", userLogin).Return([]models.Withdrawal{}, nil).Once()
			},
			want: want{
				code: http.StatusNoContent,
			},
		},
		{
			name:             "wrong method",
			method:           http.MethodPost,
			userLogin:        userLogin,
			servicesMockCall: func() {},
			want: want{
				code: http.StatusMethodNotAllowed,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.servicesMockCall()

			req := httptest.NewRequest(test.method, "/api/user/withdrawals", nil)
			if test.userLogin != "" {
				ctx := context.WithValue(req.Context(), models.LoginKey{}, test.userLogin)
				req = req.WithContext(ctx)
			}

			recorder := httptest.NewRecorder()
			r.ServeHTTP(recorder, req)

			res := recorder.Result()
			defer res.Body.Close()

			assert.Equal(t, test.want.code, res.StatusCode)
			switch res.StatusCode {
			case http.StatusOK:
				var resp []WithdrawalResp
				require.NoError(t, json.NewDecoder(res.Body).Decode(&resp))
				assert.Equal(t, test.want.resp, resp)
			case http.StatusNoContent:
				assert.Empty(t, recorder.Body, "response body should be empty")
			}
		})
	}
}
