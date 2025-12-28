package accrual

type Status = string

const (
	Registered Status = "REGISTERED"
	Processing Status = "PROCESSING"
	Invalid    Status = "INVALID"
	Processed  Status = "PROCESSED"
)
