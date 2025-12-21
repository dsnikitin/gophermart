package models

import (
	"encoding/json"
	"time"

	"github.com/dsnikitin/gophermart/internal/pkg/consts/status"
)

type Order struct {
	Number     string             `json:"number"`
	Status     status.OrderStatus `json:"status"`
	UploadedAt time.Time          `json:"uploaded_at"`
	Accrual    float64            `json:"accrual,omitempty"`
	UserLogin  string             `json:"-"`
}

func (m *Order) ScanFields() []any {
	return []any{&m.Number, &m.Status, &m.UploadedAt, &m.Accrual, &m.UserLogin}
}

func (r Order) MarshalJSON() ([]byte, error) {
	type Alias Order
	resp := &struct {
		Alias
		UploadedAt string `json:"uploaded_at"`
	}{
		Alias:      Alias(r),
		UploadedAt: r.UploadedAt.Truncate(time.Second).Format(time.RFC3339),
	}

	return json.Marshal(resp)
}
