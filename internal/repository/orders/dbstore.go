package orders

import (
	"context"
	"database/sql"
	"errors"

	"github.com/go-rfe/logging/log"
	"github.com/go-rfe/loyalty-system/internal/models"
	"github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4/stdlib" // init postgresql driver
)

const (
	pgErrCodeUniqueViolation = "23505"
)

type DBStore struct {
	connection *sql.DB
}

func NewDBStore(connection *sql.DB) *DBStore {
	db := DBStore{
		connection: connection,
	}

	return &db
}

func (db *DBStore) CreateOrder(ctx context.Context, login string, order string) error {
	var pgErr *pgconn.PgError

	_, err := db.connection.ExecContext(ctx,
		"INSERT INTO orders (number, login) VALUES ($1, $2)",
		order, login)

	switch {
	case err != nil && errors.As(err, &pgErr) && pgErr.Code == pgErrCodeUniqueViolation:
		var existingOrder string
		var orderUser string

		row := db.connection.QueryRowContext(ctx,
			"SELECT number, login FROM orders WHERE number = $1", order)

		err := row.Scan(&existingOrder, &orderUser)
		if err != nil {
			return err
		}

		if login != orderUser {
			return ErrOtherOrderExists
		}
		return ErrOrderExists
	case err != nil:
		return err
	}

	return nil
}

func (db *DBStore) UpdateOrder(ctx context.Context, order *models.Order) error {
	_, err := db.connection.ExecContext(ctx, "UPDATE orders set accrual = $1, status = $2 WHERE number = $3",
		order.Accrual, order.Status, order.Number)

	return err
}

func (db *DBStore) GetOrders(ctx context.Context, login string) ([]models.Order, error) {
	orders := make([]models.Order, 0)

	ordersRows, err := db.connection.QueryContext(ctx,
		"SELECT number,accrual,status,uploaded_at FROM orders WHERE login = $1 AND withdraw IS NULL", login)

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
		var order models.Order
		err = ordersRows.Scan(&order.Number, &order.Accrual, &order.Status, &order.UploadedAt)
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

func (db *DBStore) GetProcessedOrders(ctx context.Context, login string) ([]models.Order, error) {
	orders := make([]models.Order, 0)

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
		var order models.Order
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

func (db *DBStore) GetUnprocessedOrders(ctx context.Context) ([]models.Order, error) {
	orders := make([]models.Order, 0)

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
		var order models.Order
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

func (db *DBStore) Withdraw(ctx context.Context, login string, withdraw *models.Withdraw) error {
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

	_, err = db.connection.ExecContext(ctx,
		"INSERT INTO orders (number, login, withdraw) VALUES ($1, $2, $3)", withdraw.Order, login, withdraw.Sum)

	return err
}

func (db *DBStore) GetWithdrawals(ctx context.Context, login string) ([]models.Withdraw, error) {
	withdrawals := make([]models.Withdraw, 0)

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
		var withdraw models.Withdraw
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
