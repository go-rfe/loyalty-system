package server

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-chi/jwtauth/v5"
	"github.com/go-rfe/logging/log"
	"github.com/go-rfe/loyalty-system/internal/repository/users"
)

type Config struct {
	ServerAddress  string `env:"RUN_ADDRESS"`
	DatabaseURI    string `env:"DATABASE_URI"`
	AuthToken      []byte `env:"AUTH_TOKEN"`
	AccrualAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"`
	LogLevel       string `env:"LOG_LEVEL"`

	UserStore users.Store
}

type LoyaltyServer struct {
	Cfg      *Config
	context  context.Context
	listener *http.Server
}

func (s *LoyaltyServer) Start(ctx context.Context) {
	serverContext, serverCancel := context.WithCancel(ctx)
	s.context = serverContext

	if len(s.Cfg.AuthToken) == 0 {
		s.Cfg.AuthToken = getRandomToken()
	}

	err := initUserStore(serverContext, s.Cfg)
	if err != nil {
		log.Fatal().Msgf("Failed to initialize user store: %q", err)
	}

	go s.startListener()
	log.Info().Msgf("Start listener on %s", s.Cfg.ServerAddress)

	log.Info().Msgf("%s signal received, graceful shutdown the server", <-getSignalChannel())
	s.stopListener()

	serverCancel()
}

func (s *LoyaltyServer) AuthToken() *jwtauth.JWTAuth {
	return jwtauth.New("HS256", s.Cfg.AuthToken, nil)
}

func getSignalChannel() chan os.Signal {
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)

	return signalChannel
}
