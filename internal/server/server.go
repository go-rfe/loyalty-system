package server

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/jwtauth/v5"
	"github.com/go-rfe/logging/log"
	"github.com/go-rfe/loyalty-system/internal/repository/orders"
	"github.com/go-rfe/loyalty-system/internal/repository/users"
)

type Config struct {
	ServerAddress  string        `env:"RUN_ADDRESS"`
	DatabaseURI    string        `env:"DATABASE_URI"`
	AuthToken      []byte        `env:"AUTH_TOKEN"`
	AccrualAddress string        `env:"ACCRUAL_SYSTEM_ADDRESS"`
	AccrualScheme  string        `env:"ACCRUAL_SYSTEM_SCHEME" envDefault:"http"`
	PollInterval   time.Duration `env:"POLL_INTERVAL" envDefault:"10s"`

	LogLevel string `env:"LOG_LEVEL"`

	UserStore   users.Store
	OrdersStore orders.Store

	jwtToken *jwtauth.JWTAuth
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

	s.Cfg.jwtToken = jwtauth.New("HS256", s.Cfg.AuthToken, s.Cfg.AuthToken)

	closeUsersStore := initUsersStore(s.Cfg)
	closeOrdersStore := initOrdersStore(s.Cfg)

	pollWorker := PollerWorker{Cfg: PollerConfig{
		AccrualAddress: s.Cfg.AccrualAddress,
		AccrualScheme:  s.Cfg.AccrualScheme,
		PollInterval:   s.Cfg.PollInterval,
	}}

	pollContext, cancelPoller := context.WithCancel(ctx)
	go pollWorker.Run(pollContext, s.Cfg.OrdersStore)

	go s.startListener()
	log.Info().Msgf("Start listener on %s", s.Cfg.ServerAddress)

	log.Info().Msgf("%s signal received, graceful shutdown the server", <-getSignalChannel())
	cancelPoller()
	s.stopListener()

	if err := closeUsersStore(); err != nil {
		log.Error().Err(err).Msg("Some error occurred while users store close")
	}
	if err := closeOrdersStore(); err != nil {
		log.Error().Err(err).Msg("Some error occurred while orders store close")
	}

	serverCancel()
}

func (s *LoyaltyServer) AuthToken() *jwtauth.JWTAuth {
	return s.Cfg.jwtToken
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
