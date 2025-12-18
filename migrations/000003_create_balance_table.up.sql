CREATE TABLE IF NOT EXISTS gophermart.withdrawals (
    order_number TEXT NOT NULL PRIMARY KEY,
    user_login TEXT NOT NULL REFERENCES gophermart.users (login) ON DELETE RESTRICT ON UPDATE CASCADE,
    amount BIGINT NOT NULL CHECK (amount > 0),
    processed_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_orders_user_login ON gophermart.orders(user_login);
CREATE INDEX IF NOT EXISTS idx_withdrawals_user_login ON gophermart.withdrawals(user_login);
