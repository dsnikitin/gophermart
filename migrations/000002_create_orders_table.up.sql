CREATE TYPE gophermart.order_status AS ENUM (
    'NEW',
    'PROCESSING',
    'INVALID',
    'PROCESSED'
);

CREATE TABLE IF NOT EXISTS gophermart.orders (
    number TEXT PRIMARY KEY,
    status gophermart.order_status NOT NULL DEFAULT 'NEW',
    accrual NUMERIC(17, 2) NOT NULL DEFAULT 0 CHECK (accrual >= 0),
    uploaded_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    user_login TEXT REFERENCES gophermart.users (login)
);

CREATE INDEX idx_orders_user_login_uploaded_at ON gophermart.orders (user_login, uploaded_at DESC);
