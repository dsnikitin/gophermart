CREATE TABLE IF NOT EXISTS gophermart.withdrawals (
    order_number TEXT NOT NULL,
    user_login TEXT NOT NULL REFERENCES gophermart.users (login) ON DELETE RESTRICT ON UPDATE CASCADE,
    amount NUMERIC(17,2) NOT NULL CHECK (amount > 0),
    processed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_orders_user_login ON gophermart.orders(user_login);
CREATE INDEX IF NOT EXISTS idx_withdrawals_user_login ON gophermart.withdrawals(user_login);
CREATE INDEX IF NOT EXISTS idx_withdrawals_user_login_processed_at ON gophermart.withdrawals (user_login, processed_at DESC);
