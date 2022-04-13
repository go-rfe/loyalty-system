package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-rfe/logging/log"
	"github.com/go-rfe/loyalty-system/internal/repository/orders"
)

func BalanceHandler(ordersStore orders.Store) func(r chi.Router) {
	return func(r chi.Router) {
		r.Get("/", getBalance(ordersStore))
		r.Get("/withdrawals", getWithdrawals(ordersStore))
		r.Post("/withdraw", withdraw(ordersStore))
	}
}

func getBalance(ordersStore orders.Store) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		requestContext, requestCancel := context.WithTimeout(r.Context(), requestTimeout)
		defer requestCancel()

		login, err := getLoginFromRequest(r)
		if err != nil {
			http.Error(
				w,
				fmt.Sprintf("couldn't get user from token: %q", err),
				http.StatusInternalServerError,
			)

			return
		}

		balance, err := ordersStore.GetBalance(requestContext, login)
		if err != nil {
			http.Error(
				w,
				fmt.Sprintf("couldn't get balance for %s: %q", login, err),
				http.StatusInternalServerError,
			)

			return
		}

		w.Header().Set("Content-Type", "application/json")
		err = orders.Encode(balance, w)
		if err != nil {
			log.Error().Err(err).Msg("Cannot send request")
		}
	}
}

func getWithdrawals(ordersStore orders.Store) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		requestContext, requestCancel := context.WithTimeout(r.Context(), requestTimeout)
		defer requestCancel()

		login, err := getLoginFromRequest(r)
		if err != nil {
			http.Error(
				w,
				fmt.Sprintf("couldn't get user from token: %q", err),
				http.StatusInternalServerError,
			)

			return
		}

		withdrawals, err := ordersStore.GetWithdrawals(requestContext, login)
		if err != nil {
			http.Error(
				w,
				fmt.Sprintf("couldn't get balance for %s: %q", login, err),
				http.StatusInternalServerError,
			)

			return
		}

		w.Header().Set("Content-Type", "application/json")
		err = orders.Encode(withdrawals, w)
		if err != nil {
			log.Error().Err(err).Msg("Cannot send request")
		}
	}
}

func withdraw(ordersStore orders.Store) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		requestContext, requestCancel := context.WithTimeout(r.Context(), requestTimeout)
		defer requestCancel()

		login, err := getLoginFromRequest(r)
		if err != nil {
			http.Error(
				w,
				fmt.Sprintf("couldn't get user from token: %q", err),
				http.StatusInternalServerError,
			)

			return
		}

		var withdraw orders.Withdraw
		err = json.NewDecoder(r.Body).Decode(&withdraw)
		if err != nil {
			http.Error(w, fmt.Sprintf("Cannot decode provided data: %q", err), http.StatusBadRequest)

			return
		}

		if err := orders.Validate(withdraw.Order); err != nil {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)

			return
		}

		err = ordersStore.Withdraw(requestContext, login, &withdraw)
		if err != nil && errors.Is(err, orders.ErrInsufficientBalance) {
			http.Error(w, err.Error(), http.StatusPaymentRequired)

			return
		}
		if err != nil {
			http.Error(
				w,
				fmt.Sprintf("couldn't update balance for %s: %q", login, err),
				http.StatusInternalServerError,
			)

			return
		}
	}
}
