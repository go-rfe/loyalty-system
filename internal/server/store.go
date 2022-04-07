package server

import (
	"context"

	"github.com/go-rfe/logging/log"

	"github.com/go-rfe/loyalty-system/internal/repository/users"
)

func initUserStore(ctx context.Context, config *Config) error {
	switch {
	case config.DatabaseURI != "":
		return nil
	default:
		log.Info().Msg("Using memory storage")
		config.UserStore = users.NewInMemoryStore()
	}

	return nil
}
