package server

import (
	"compress/gzip"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-rfe/logging"
	"github.com/go-rfe/logging/log"
)

func (s *LoyaltyServer) startListener() {
	mux := chi.NewRouter()

	mux.Use(logging.HTTPRequestLogger(s.Cfg.LogLevel))
	mux.Use(middleware.RequestID)
	mux.Use(middleware.RealIP)
	mux.Use(middleware.Recoverer)

	compressor := middleware.NewCompressor(gzip.BestCompression)
	mux.Use(compressor.Handler)

	RegisterPublicHandlers(mux, s.Cfg.UserStore, s.AuthToken())
	RegisterPrivateHandlers(mux, s.Cfg.UserStore, s.AuthToken())

	httpServer := &http.Server{
		Addr:    s.Cfg.ServerAddress,
		Handler: mux,
	}

	s.listener = httpServer

	log.Info().Msgf("%v", s.listener.ListenAndServe())
}

func (s *LoyaltyServer) stopListener() {
	err := s.listener.Shutdown(s.context)
	if err != nil {
		log.Info().Msgf("HTTP server ListenAndServe shut down: %v", err)
	}
}
