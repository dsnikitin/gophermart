package models

import "github.com/dsnikitin/gophermart/internal/pkg/consts/accrual"

type Accrual struct {
	OrderNumber string         `json:"number"`
	Status      accrual.Status `json:"status"`
	Accrual     float64        `json:"accrual"`
}
