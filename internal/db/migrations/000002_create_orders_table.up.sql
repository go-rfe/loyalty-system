CREATE TABLE IF NOT EXISTS orders(
    number BIGINT PRIMARY KEY,
    login  VARCHAR (50) REFERENCES users(login),
    status VARCHAR (50) DEFAULT 'NEW',
    accrual REAL DEFAULT 0,
    uploaded_at TIMESTAMP
);
