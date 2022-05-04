package users

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4/stdlib" // init postgresql driver

	"github.com/go-rfe/loyalty-system/internal/models"
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

func (db *DBStore) CreateUser(ctx context.Context, login string, password string) error {
	var pgErr *pgconn.PgError

	_, err := db.connection.ExecContext(ctx,
		"INSERT INTO users (login, password) VALUES ($1, $2)",
		login, password)

	if err != nil && errors.As(err, &pgErr) && pgErr.Code == pgErrCodeUniqueViolation {
		return ErrUserExists
	}

	return err
}

func (db *DBStore) ValidateUser(ctx context.Context, login string, password string) error {
	var userPassword string
	row := db.connection.QueryRowContext(ctx,
		"SELECT password FROM users WHERE login = $1", login)

	err := row.Scan(&userPassword)
	if !errors.Is(err, nil) && !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	if errors.Is(err, sql.ErrNoRows) {
		return ErrUserNotFound
	}

	user := &models.User{
		Login:    login,
		Password: userPassword,
	}

	if err := user.CheckPassword(password); err != nil {
		return models.ErrInvalidPassword
	}

	return err
}

func (db *DBStore) Close() error {
	return db.connection.Close()
}
