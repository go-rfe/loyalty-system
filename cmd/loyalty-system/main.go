package main

import (
	"context"

	"github.com/caarlos0/env/v6"
	"github.com/go-rfe/logging"
	"github.com/go-rfe/logging/log"
	"github.com/go-rfe/loyalty-system/cmd"
	"github.com/go-rfe/loyalty-system/internal/server"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatal().Err(err).Msg("Failed to parse command line arguments")
	}

	LoyaltyServerConfig := server.Config{
		ServerAddress: cmd.ServerAddress,
		LogLevel:      cmd.LogLevel,
		DatabaseURI:   cmd.DatabaseURI,
	}
	if err := env.Parse(&LoyaltyServerConfig); err != nil {
		log.Fatal().Err(err).Msg("Failed to parse environment variables")
	}

	logging.Level(LoyaltyServerConfig.LogLevel)

	loyaltyServer := server.LoyaltyServer{Cfg: &LoyaltyServerConfig}

	loyaltyServer.Start(context.Background())
}
