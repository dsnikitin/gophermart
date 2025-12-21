package models

type Balance struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

func (m *Balance) ScanFields() []any {
	return []any{&m.Current, &m.Withdrawn}
}
