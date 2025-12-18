CREATE TYPE gophermart.order_status AS ENUM (
    'NEW',
    'PROCESSING',
    'INVALID',
    'PROCESSED'
);

CREATE TABLE IF NOT EXISTS gophermart.orders (
    number TEXT PRIMARY KEY,
    status gophermart.order_status NOT NULL,
    accrual BIGINT NOT NULL CHECK (accrual >= 0),
    uploaded_at TIMESTAMPTZ NOT NULL,
    user_login TEXT REFERENCES gophermart.users (login)
);

CREATE INDEX idx_orders_user_uploaded_at ON gophermart.orders (user_login, uploaded_at DESC);