CREATE TABLE IF NOT EXISTS orders(
    number BIGINT PRIMARY KEY,
    login  VARCHAR (50) REFERENCES users(login),
    status VARCHAR (50) DEFAULT 'NEW',
    accrual DECIMAL DEFAULT NULL,
    withdraw DECIMAL DEFAULT NULL,
    uploaded_at TIMESTAMP DEFAULT now()
);
