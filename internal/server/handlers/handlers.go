package handlers

import (
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"github.com/go-rfe/loyalty-system/internal/repository/orders"
	"github.com/go-rfe/loyalty-system/internal/repository/users"
)

const (
	requestTimeout = 1 * time.Second
)

func RegisterPublicHandlers(mux *chi.Mux, userStore users.Store, auth *jwtauth.JWTAuth) {
	mux.Group(func(r chi.Router) {
		r.Route("/api/user/register", UserRegisterHandler(userStore, auth))
		r.Route("/api/user/login", UserLoginHandler(userStore, auth))
	})
}

func RegisterPrivateHandlers(mux *chi.Mux, ordersStore orders.Store, auth *jwtauth.JWTAuth) {
	mux.Group(func(r chi.Router) {
		r.Use(jwtauth.Verifier(auth))
		r.Use(jwtauth.Authenticator)

		r.Route("/api/user/orders", OrdersHandler(ordersStore))
		r.Route("/api/user/balance", BalanceHandler(ordersStore))
	})
}
