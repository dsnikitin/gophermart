package models

import (
	"time"

	"github.com/dsnikitin/gophermart/internal/pkg/consts/status"
)

type Order struct {
	Number     string             `json:"number"`
	Status     status.OrderStatus `json:"status"`
	Accrual    int64              `json:"accrual,omitempty"`
	UploadedAt time.Time          `json:"uploaded_at"` // RFC3339
	UserLogin  string             `json:"-"`
}

func (m *Order) ScanFields() []any {
	return []any{&m.Number, &m.Status, &m.Accrual, &m.UploadedAt, &m.UserLogin}
}
