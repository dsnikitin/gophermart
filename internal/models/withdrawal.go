package models

import (
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type WithdrawRequest struct {
	Order string  `json:"order"`
	Sum   float64 `json:"sum"`
}

func (m *WithdrawRequest) Validate() error {
	return validation.ValidateStruct(m,
		validation.Field(&m.Order, validation.Required),
		validation.Field(&m.Sum, validation.Required, validation.Min(0.01)),
	)
}

type WithdrawalDB struct {
	OrderNumber string
	Amount      int64
	ProcessedAt time.Time
}

func (m *WithdrawalDB) ScanFields() []any {
	return []any{&m.OrderNumber, &m.Amount, &m.ProcessedAt}
}

func (m *WithdrawalDB) ToWithdrawalResponse() WithdrawalResponse {
	return WithdrawalResponse{
		Order:       m.OrderNumber,
		Sum:         float64(m.Amount / 100),
		ProcessedAt: m.ProcessedAt.Truncate(time.Second),
	}
}

type WithdrawalResponse struct {
	Order       string    `json:"order"`
	Sum         float64   `json:"sum"`
	ProcessedAt time.Time `json:"processed_at"` // с точностью до секунд
}
