package models

import (
	"time"

	"github.com/dsnikitin/gophermart/internal/pkg/consts/status"
)

type OrderDB struct {
	Number     string
	Status     status.OrderStatus
	UploadedAt time.Time
	Accrual    int64
	UserLogin  string
}

func (m *OrderDB) ScanFields() []any {
	return []any{&m.Number, &m.Status, &m.UploadedAt, &m.Accrual, &m.UserLogin}
}

func (m *OrderDB) ToOrderResponse() OrderResponse {
	return OrderResponse{
		Number:     m.Number,
		Status:     m.Status,
		UploadedAt: m.UploadedAt.Truncate(time.Second),
		Accrual:    float64(m.Accrual / 100),
	}
}

type OrderResponse struct {
	Number     string             `json:"number"`
	Status     status.OrderStatus `json:"status"`
	UploadedAt time.Time          `json:"uploaded_at"` // с точностью до секунд
	Accrual    float64            `json:"accrual,omitempty"`
}
