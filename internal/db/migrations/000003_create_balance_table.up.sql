CREATE TABLE IF NOT EXISTS balance(
    login  VARCHAR (50) REFERENCES users(login),
    order_number INTEGER REFERENCES orders(number),
    sum REAL,
    processed_at TIMESTAMP
);
