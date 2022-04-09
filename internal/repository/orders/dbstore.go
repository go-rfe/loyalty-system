package orders

import (
	"context"
	"database/sql"
	"errors"

	_ "github.com/jackc/pgx/v4/stdlib" // init postgresql driver

	"github.com/go-rfe/logging/log"
)

const (
	psqlDriverName = "pgx"
)

type DBStore struct {
	connection *sql.DB
}

func NewDBStore(databaseDSN string) (*DBStore, error) {
	var db DBStore

	conn, err := sql.Open(psqlDriverName, databaseDSN)
	if err != nil {
		return nil, err
	}

	db = DBStore{
		connection: conn,
	}

	return &db, nil
}

func (db *DBStore) CreateOrder(ctx context.Context, login string, order *Order) error {
	var existingOrder int64
	var orderLogin string
	row := db.connection.QueryRowContext(ctx,
		"SELECT number, login FROM orders WHERE number = $1", order.Number)

	err := row.Scan(&existingOrder, &orderLogin)
	if !errors.Is(err, nil) && !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	if err == nil && login != orderLogin {
		return ErrOtherOrderExists
	}
	if login == orderLogin {
		return ErrOrderExists
	}

	_, err = db.connection.ExecContext(ctx,
		"INSERT INTO orders (number, login, uploaded_at) VALUES ($1, $2, $3)", order.Number, login, order.UploadedAt)

	return err
}

func (db *DBStore) GetOrders(ctx context.Context, login string) ([]Order, error) {
	orders := make([]Order, 0)

	ordersRows, err := db.connection.QueryContext(ctx,
		"SELECT number,status,accrual,uploaded_at FROM orders WHERE login = $1", login)

	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			log.Error().Err(err).Msgf("Couldn't close rows")
		}
	}(ordersRows)

	for ordersRows.Next() {
		var order Order
		err = ordersRows.Scan(&order.Number, &order.Status, &order.Accrual, &order.UploadedAt)
		if err != nil {
			return nil, err
		}

		orders = append(orders, order)
	}

	err = ordersRows.Err()
	if err != nil {
		return nil, err
	}

	return orders, nil
}

func (db *DBStore) Close() error {
	return db.connection.Close()
}
