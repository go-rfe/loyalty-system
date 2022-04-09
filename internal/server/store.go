package server

import (
	"github.com/go-rfe/logging/log"
	"github.com/go-rfe/loyalty-system/internal/repository/orders"

	"github.com/go-rfe/loyalty-system/internal/repository/users"
)

func initUsersStore(config *Config) func() error {
	switch {
	case config.DatabaseURI != "":
		userStore, err := users.NewDBStore(config.DatabaseURI)
		if err != nil {
			log.Fatal().Msgf("Couldn't connect to database: %q", err)
		}

		config.UserStore = userStore

		log.Info().Msg("Using Database storage")

		return func() error {
			return userStore.Close()
		}
	default:
		log.Info().Msg("Using memory storage")
		config.UserStore = users.NewInMemoryStore()

		return func() error {
			return nil
		}
	}
}

func initOrdersStore(config *Config) func() error {
	ordersStore, err := orders.NewDBStore(config.DatabaseURI)
	if err != nil {
		log.Fatal().Msgf("Couldn't connect to database: %q", err)
	}

	config.OrdersStore = ordersStore

	log.Info().Msg("Using Database storage")

	return func() error {
		return ordersStore.Close()
	}
}
