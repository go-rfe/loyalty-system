package server

import (
	"database/sql"

	"github.com/go-rfe/logging/log"
	"github.com/go-rfe/loyalty-system/internal/repository/orders"

	"github.com/go-rfe/loyalty-system/internal/repository/users"
)

const (
	psqlDriverName = "pgx"
)

func initStore(config *Config) (func() error, func() error) {
	conn, err := sql.Open(psqlDriverName, config.DatabaseURI)
	if err != nil {
		log.Fatal().Err(err).Msg("Couldn't create database connection")
	}

	userStore := users.NewDBStore(conn)
	config.UserStore = userStore
	log.Info().Msg("Using Database for user storage")

	ordersStore := orders.NewDBStore(conn)
	config.OrdersStore = ordersStore
	log.Info().Msg("Using Database for orders storage")

	return userStore.Close, ordersStore.Close
}
