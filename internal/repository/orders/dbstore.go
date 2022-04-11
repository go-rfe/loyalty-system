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

func (db *DBStore) CreateOrder(ctx context.Context, login string, order string) error {
	var existingOrder int64
	var orderLogin string
	row := db.connection.QueryRowContext(ctx,
		"SELECT number, login FROM orders WHERE number = $1", order)

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
		"INSERT INTO orders (number, login) VALUES ($1, $2)", order, login)

	return err
}

func (db *DBStore) UpdateOrder(ctx context.Context, order *Order) error {
	withoutAccrualStatuses := map[string]struct{}{
		"INVALID":    {},
		"REGISTERED": {},
		"PROCESSING": {},
	}

	var existingOrder string
	row := db.connection.QueryRowContext(ctx,
		"SELECT number FROM orders WHERE number = $1", order.Number)

	err := row.Scan(&existingOrder)
	if err != nil {
		return err
	}

	if _, ok := withoutAccrualStatuses[order.Status]; ok {
		_, err = db.connection.ExecContext(ctx, "UPDATE orders set status = $1 WHERE number = $2",
			order.Status, order.Number)
	} else {
		_, err = db.connection.ExecContext(ctx, "UPDATE orders set accrual = $1, status = $2 WHERE number = $3",
			order.Accrual, order.Status, order.Number)
	}

	return err
}

func (db *DBStore) GetOrders(ctx context.Context, login string) ([]Order, error) {
	orders := make([]Order, 0)

	processedOrders, err := db.getProcessedOrders(ctx, login)
	if err != nil {
		return nil, err
	}
	otherOrders, err := db.getOtherOrders(ctx, login)
	if err != nil {
		return nil, err
	}

	orders = append(orders, processedOrders...)
	orders = append(orders, otherOrders...)

	return orders, nil
}

func (db *DBStore) getProcessedOrders(ctx context.Context, login string) ([]Order, error) {
	orders := make([]Order, 0)

	ordersRows, err := db.connection.QueryContext(ctx,
		"SELECT number,status,accrual,uploaded_at FROM orders WHERE login = $1 AND status = 'PROCESSED'", login)

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

func (db *DBStore) getOtherOrders(ctx context.Context, login string) ([]Order, error) {
	orders := make([]Order, 0)

	ordersRows, err := db.connection.QueryContext(ctx,
		"SELECT number,status,uploaded_at FROM orders WHERE login = $1 AND withdraw IS NULL AND NOT status = 'PROCESSED'", login)

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
		err = ordersRows.Scan(&order.Number, &order.Status, &order.UploadedAt)
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

func (db *DBStore) GetUnprocessedOrders(ctx context.Context) ([]Order, error) {
	orders := make([]Order, 0)

	ordersRows, err := db.connection.QueryContext(ctx,
		"SELECT number FROM orders WHERE status = 'NEW'")

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
		err = ordersRows.Scan(&order.Number)
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

func (db *DBStore) GetBalance(ctx context.Context, login string) (*Balance, error) {
	var (
		balance   Balance
		withdrawn float32
		accrual   float32
	)

	processedOrders, err := db.getProcessedOrders(ctx, login)
	if err != nil {
		return nil, err
	}

	withdrawals, err := db.GetWithdrawals(ctx, login)
	if err != nil {
		return nil, err
	}

	for _, order := range processedOrders {
		accrual += order.Accrual
	}

	for _, withdraw := range withdrawals {
		withdrawn += withdraw.Sum
		accrual -= withdraw.Sum
	}

	balance = Balance{
		Current:   accrual,
		Withdrawn: withdrawn,
	}

	return &balance, nil
}

func (db *DBStore) Withdraw(ctx context.Context, login string, withdraw *Withdraw) error {
	var existingOrder int64
	var orderLogin string
	row := db.connection.QueryRowContext(ctx,
		"SELECT number, login FROM orders WHERE number = $1", withdraw.Order)

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

	balance, err := db.GetBalance(ctx, login)
	if err != nil {
		return err
	}

	if balance.Current-withdraw.Sum < 0 {
		return ErrInsufficientBalance
	}

	_, err = db.connection.ExecContext(ctx,
		"INSERT INTO orders (number, login, withdraw) VALUES ($1, $2, $3)", withdraw.Order, login, withdraw.Sum)

	return err
}

func (db *DBStore) GetWithdrawals(ctx context.Context, login string) ([]Withdraw, error) {
	withdrawals := make([]Withdraw, 0)

	withdrawalsRows, err := db.connection.QueryContext(ctx,
		"SELECT number,withdraw,uploaded_at FROM orders WHERE login = $1 AND withdraw IS NOT NULL", login)

	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			log.Error().Err(err).Msgf("Couldn't close rows")
		}
	}(withdrawalsRows)

	for withdrawalsRows.Next() {
		var withdraw Withdraw
		err = withdrawalsRows.Scan(&withdraw.Order, &withdraw.Sum, &withdraw.ProcessedAt)
		if err != nil {
			return nil, err
		}

		withdrawals = append(withdrawals, withdraw)
	}

	err = withdrawalsRows.Err()
	if err != nil {
		return nil, err
	}

	return withdrawals, nil
}

func (db *DBStore) Close() error {
	return db.connection.Close()
}
