package service

type Service struct {
	User    *UserService
	Order   *OrderService
	Balance *BalanceService
}

type Repository struct {
	User       UserRepository
	Order      OrderRepository
	Balance    BalanceRepository
	TxProvider TransactionProvider
}

func New(r *Repository) *Service {
	return &Service{
		User:    NewUser(r.User),
		Order:   NewOrder(r.Order),
		Balance: NewBalance(r.Balance, r.TxProvider),
	}
}
