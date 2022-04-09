package users

import (
	"context"
	"database/sql"
	"errors"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	bindata "github.com/golang-migrate/migrate/v4/source/go_bindata"

	_ "github.com/jackc/pgx/v4/stdlib" // init postgresql driver

	"github.com/go-rfe/loyalty-system/internal/db/migrations"
)

const (
	psqlDriverName      = "pgx"
	migrationSourceName = "go-bindata"
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

	if err := db.migrate(); err != nil {
		return nil, err
	}

	return &db, nil
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

func (db *DBStore) migrate() error {
	data := bindata.Resource(migrations.AssetNames(), migrations.Asset)

	sourceDriver, err := bindata.WithInstance(data)
	if err != nil {
		return err
	}

	dbDriver, err := postgres.WithInstance(db.connection, &postgres.Config{})
	if err != nil {
		return err
	}

	migration, err := migrate.NewWithInstance(migrationSourceName, sourceDriver, psqlDriverName, dbDriver)
	if err != nil {
		return err
	}

	if err := migration.Up(); !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	return nil
}
