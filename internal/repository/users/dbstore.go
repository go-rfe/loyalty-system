package users

import (
	"context"
	"database/sql"
	"errors"

	_ "github.com/jackc/pgx/v4/stdlib" // init postgresql driver
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
	var existingUser string
	row := db.connection.QueryRowContext(ctx,
		"SELECT login FROM users WHERE login = $1", login)

	err := row.Scan(&existingUser)
	if !errors.Is(err, nil) && !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return ErrUserExists
	}

	user := &User{
		Login: login,
	}

	if err := user.SetPassword(password); err != nil {
		return err
	}

	_, err = db.connection.ExecContext(ctx,
		"INSERT INTO users (login, password) VALUES ($1, $2)",
		user.Login, user.Password)

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

	user := &User{
		Login:    login,
		Password: userPassword,
	}

	if err := user.CheckPassword(password); err != nil {
		return ErrInvalidPassword
	}

	return err
}

func (db *DBStore) Close() error {
	return db.connection.Close()
}
