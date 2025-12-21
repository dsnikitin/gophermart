package models

import (
	"encoding/json"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type WithdrawRequest struct {
	Order string  `json:"order"`
	Sum   float64 `json:"sum"`
}

func (m WithdrawRequest) Validate() error {
	return validation.ValidateStruct(&m,
		validation.Field(&m.Order, validation.Required),
		validation.Field(&m.Sum, validation.Required, validation.Min(0.01)),
	)
}

type Withdrawal struct {
	OrderNumber string    `json:"order"`
	Sum         float64   `json:"sum"`
	ProcessedAt time.Time `json:"processed_at"`
}

func (m *Withdrawal) ScanFields() []any {
	return []any{&m.OrderNumber, &m.Sum, &m.ProcessedAt}
}

func (r Withdrawal) MarshalJSON() ([]byte, error) {
	type Alias Withdrawal
	resp := &struct {
		Alias
		ProcessedAt string `json:"processed_at"`
	}{
		Alias:       Alias(r),
		ProcessedAt: r.ProcessedAt.Truncate(time.Second).Format(time.RFC3339),
	}

	return json.Marshal(resp)
}
