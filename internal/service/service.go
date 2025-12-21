package service

import "github.com/dsnikitin/gophermart/internal/config"

type Service struct {
	User    *UserService
	Order   *OrderService
	Balance *BalanceService
	Accrual *AccrualService
}

type Dependencies struct {
	User      UserRepository
	Order     OrderRepository
	Balance   BalanceRepository
	BalanceTx BalanceTxProvider
	Accrual   AccrualRepository
	AccrualTx AccrualTxProvider
}

func New(cfg *config.Config, d *Dependencies) *Service {
	return &Service{
		User:    NewUserService(d.User),
		Order:   NewOrderService(d.Order),
		Balance: NewBalanceService(d.Balance, d.BalanceTx),
		Accrual: NewAccrualService(cfg.AccrualSystemAddr, d.Accrual, d.AccrualTx),
	}
}
