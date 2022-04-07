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

	logging.Level(cmd.LogLevel)

	LoyaltyServerConfig := server.Config{
		ServerAddress: cmd.ServerAddress,
		LogLevel:      cmd.LogLevel,
	}
	if err := env.Parse(&LoyaltyServerConfig); err != nil {
		log.Fatal().Err(err).Msg("Failed to parse environment variables")
	}

	loyaltyServer := server.LoyaltyServer{Cfg: &LoyaltyServerConfig}

	loyaltyServer.Start(context.Background())
}
