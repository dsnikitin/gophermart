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
	"github.com/dsnikitin/gophermart/internal/pkg/consts/order"
	"github.com/dsnikitin/gophermart/internal/pkg/errx"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockOrderService struct {
	mock.Mock
}

func (m *MockOrderService) UploadOrder(ctx context.Context, login, orderNumber string) error {
	args := m.Called(login, orderNumber)
	return args.Error(0)
}

func (m *MockOrderService) GetOrders(ctx context.Context, login string) ([]models.Order, error) {
	args := m.Called(login)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Order), args.Error(1)
}

type MockAccrualService struct {
	mock.Mock
}

func (m *MockAccrualService) NotifyOrderUploaded() {}

func TestUserHandler_UploadOrder(t *testing.T) {
	userLogin := "user1"
	orderNumber := "12345678903"

	type headers struct {
		contentType string
	}

	type want struct {
		code    int
		headers headers
		errMsg  string
	}

	s := new(MockOrderService)
	h := handler.New(&config.Config{}, &handler.Service{Order: s, Accrual: new(MockAccrualService)})

	r := chi.NewRouter()
	r.Post("/api/user/orders", h.Order.UploadOrder)

	tests := []struct {
		name             string
		method           string
		userLogin        string
		orderNumber      string
		servicesMockCall func()
		want             want
	}{
		{
			name:        "positive",
			method:      http.MethodPost,
			userLogin:   userLogin,
			orderNumber: orderNumber,
			servicesMockCall: func() {
				s.On("UploadOrder", userLogin, orderNumber).Return(nil).Once()
			},
			want: want{
				code: http.StatusAccepted,
			},
		},
		{
			name:             "wrong method",
			method:           http.MethodGet,
			userLogin:        userLogin,
			orderNumber:      orderNumber,
			servicesMockCall: func() {},
			want: want{
				code: http.StatusMethodNotAllowed,
			},
		},
		{
			name:             "empty request body",
			method:           http.MethodPost,
			userLogin:        userLogin,
			orderNumber:      "",
			servicesMockCall: func() {},
			want: want{
				code:    http.StatusBadRequest,
				headers: headers{contentType: "text/plain; charset=utf-8"},
				errMsg:  "empty request body\n",
			},
		},
		{
			name:        "already accepted",
			method:      http.MethodPost,
			userLogin:   userLogin,
			orderNumber: orderNumber,
			servicesMockCall: func() {
				s.On("UploadOrder", userLogin, orderNumber).Return(errx.ErrAlreadyAccepted).Once()
			},
			want: want{
				code: http.StatusOK,
			},
		},
		{
			name:        "conflict",
			method:      http.MethodPost,
			userLogin:   userLogin,
			orderNumber: orderNumber,
			servicesMockCall: func() {
				s.On("UploadOrder", userLogin, orderNumber).Return(errx.ErrAlreadyExists).Once()
			},
			want: want{
				code: http.StatusConflict,
			},
		},
		{
			name:        "invalid order number - too short",
			method:      http.MethodPost,
			userLogin:   userLogin,
			orderNumber: "1",
			servicesMockCall: func() {
				s.On("UploadOrder", userLogin, "1").Return(errx.ErrInvalidOrderNumber).Once()
			},
			want: want{
				code:    http.StatusUnprocessableEntity,
				headers: headers{contentType: "text/plain; charset=utf-8"},
				errMsg:  errx.ErrInvalidOrderNumber.Error() + "\n",
			},
		},
		{
			name:        "invalid order number - has non-digits",
			method:      http.MethodPost,
			userLogin:   userLogin,
			orderNumber: "123a",
			servicesMockCall: func() {
				s.On("UploadOrder", userLogin, "123a").Return(errx.ErrInvalidOrderNumber).Once()
			},
			want: want{
				code:    http.StatusUnprocessableEntity,
				headers: headers{contentType: "text/plain; charset=utf-8"},
				errMsg:  errx.ErrInvalidOrderNumber.Error() + "\n",
			},
		},
		{
			name:        "invalid order number - all zeros",
			method:      http.MethodPost,
			userLogin:   userLogin,
			orderNumber: "000",
			servicesMockCall: func() {
				s.On("UploadOrder", userLogin, "000").Return(errx.ErrInvalidOrderNumber).Once()
			},
			want: want{
				code:    http.StatusUnprocessableEntity,
				headers: headers{contentType: "text/plain; charset=utf-8"},
				errMsg:  errx.ErrInvalidOrderNumber.Error() + "\n",
			},
		},
		{
			name:        "invalid order number - luhn failed",
			method:      http.MethodPost,
			userLogin:   userLogin,
			orderNumber: "123456789031",
			servicesMockCall: func() {
				s.On("UploadOrder", userLogin, "123456789031").Return(errx.ErrInvalidOrderNumber).Once()
			},
			want: want{
				code:    http.StatusUnprocessableEntity,
				headers: headers{contentType: "text/plain; charset=utf-8"},
				errMsg:  errx.ErrInvalidOrderNumber.Error() + "\n",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.servicesMockCall()

			req := httptest.NewRequest(test.method, "/api/user/orders", bytes.NewBufferString(test.orderNumber))
			req.Header.Add("Content-Type", "text/plain")
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
			case http.StatusOK, http.StatusAccepted, http.StatusUnauthorized, http.StatusConflict:
				assert.Empty(t, resBody, "response body should be empty")
			case http.StatusBadRequest, http.StatusUnprocessableEntity:
				assert.Equal(t, test.want.headers.contentType, res.Header.Get("Content-Type"))
				assert.Equal(t, test.want.errMsg, string(resBody))
			}
		})
	}
}

func TestUserHandler_GetOrders(t *testing.T) {
	type OrderResp struct {
		Number     string       `json:"number"`
		Status     order.Status `json:"status"`
		UploadedAt string       `json:"uploaded_at"`
		Accrual    float64      `json:"accrual"`
	}

	userLogin := "user1"
	now := time.Now()
	minuteAgo := now.Add(time.Minute)
	ordersResp := []OrderResp{
		{
			Number:     "012345678903",
			Status:     "PROCESSED",
			UploadedAt: minuteAgo.Truncate(time.Second).Format(time.RFC3339),
			Accrual:    100.5,
		},
		{
			Number:     "12345678903",
			UploadedAt: now.Truncate(time.Second).Format(time.RFC3339),
			Status:     "NEW",
		},
	}

	type headers struct {
		contentType string
	}

	type want struct {
		code    int
		headers headers
		resp    []OrderResp
	}

	s := new(MockOrderService)
	h := handler.New(&config.Config{}, &handler.Service{Order: s})

	r := chi.NewRouter()
	r.Get("/api/user/orders", h.Order.GetOrders)

	tests := []struct {
		name             string
		method           string
		userLogin        string
		servicesMockCall func()
		want             want
	}{
		{
			name:      "positive with orders",
			method:    http.MethodGet,
			userLogin: userLogin,
			servicesMockCall: func() {
				orders := []models.Order{
					{
						Number:     "012345678903",
						Status:     "PROCESSED",
						UploadedAt: minuteAgo,
						Accrual:    100.5,
						UserLogin:  userLogin,
					}, {
						Number:     "12345678903",
						Status:     "NEW",
						UploadedAt: now,
						Accrual:    0,
						UserLogin:  userLogin,
					},
				}
				s.On("GetOrders", userLogin).Return(orders, nil).Once()
			},
			want: want{
				code:    http.StatusOK,
				headers: headers{contentType: "application/json"},
				resp:    ordersResp,
			},
		},
		{
			name:      "positive no content",
			method:    http.MethodGet,
			userLogin: userLogin,
			servicesMockCall: func() {
				s.On("GetOrders", userLogin).Return([]models.Order{}, nil).Once()
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

			req := httptest.NewRequest(test.method, "/api/user/orders", nil)
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
				var resp []OrderResp
				require.NoError(t, json.NewDecoder(res.Body).Decode(&resp))
				assert.Equal(t, test.want.resp, resp)
			case http.StatusNoContent:
				assert.Empty(t, recorder.Body, "response body should be empty")
			}
		})
	}
}
