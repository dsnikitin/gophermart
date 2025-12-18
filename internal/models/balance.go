package models

type BalanceDB struct {
	Current   int64
	Withdrawn int64
}

func (m *BalanceDB) ScanFields() []any {
	return []any{&m.Current, &m.Withdrawn}
}

func (m *BalanceDB) ToBalanceResponse() BalanceResponse {
	return BalanceResponse{Current: float64(m.Current / 100), Withdrawn: float64(m.Withdrawn / 100)}
}

type BalanceResponse struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}
