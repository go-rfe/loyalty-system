package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-rfe/logging/log"
	"github.com/go-rfe/loyalty-system/internal/models"
	"github.com/go-rfe/loyalty-system/internal/repository/orders"
	"github.com/shopspring/decimal"
)

func BalanceHandler(ordersStore orders.Store) func(r chi.Router) {
	return func(r chi.Router) {
		r.Get("/", getBalanceHandler(ordersStore))
		r.Get("/withdrawals", getWithdrawalsHandler(ordersStore))
		r.Post("/withdraw", withdrawHandler(ordersStore))
	}
}

func getBalanceHandler(ordersStore orders.Store) func(w http.ResponseWriter, r *http.Request) {
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

		balance, err := getBalance(requestContext, login, ordersStore)
		if err != nil {
			http.Error(
				w,
				fmt.Sprintf("couldn't get balance for %s: %q", login, err),
				http.StatusInternalServerError,
			)

			return
		}

		w.Header().Set("Content-Type", "application/json")
		err = models.Encode(balance, w)
		if err != nil {
			log.Error().Err(err).Msg("Cannot send request")
		}
	}
}

func getWithdrawalsHandler(ordersStore orders.Store) func(w http.ResponseWriter, r *http.Request) {
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
		err = models.Encode(withdrawals, w)
		if err != nil {
			log.Error().Err(err).Msg("Cannot send request")
		}
	}
}

func withdrawHandler(ordersStore orders.Store) func(w http.ResponseWriter, r *http.Request) {
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

		var withdraw models.Withdraw
		err = json.NewDecoder(r.Body).Decode(&withdraw)
		if err != nil {
			http.Error(w, fmt.Sprintf("Cannot decode provided data: %q", err), http.StatusBadRequest)

			return
		}

		if err := models.Validate(withdraw.Order); err != nil {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)

			return
		}

		balance, err := getBalance(requestContext, login, ordersStore)
		if err != nil {
			http.Error(
				w,
				fmt.Sprintf("couldn't get balance for %s: %q", login, err),
				http.StatusInternalServerError,
			)

			return
		}

		if balance.Current.Sub(withdraw.Sum).LessThan(decimal.Zero) {
			http.Error(w, "insufficient balance", http.StatusPaymentRequired)

			return
		}

		err = ordersStore.Withdraw(requestContext, login, &withdraw)
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

func getBalance(ctx context.Context, login string, ordersStore orders.Store) (*models.Balance, error) {
	var (
		balance   models.Balance
		withdrawn decimal.Decimal
		accrual   decimal.Decimal
	)

	processedOrders, err := ordersStore.GetProcessedOrders(ctx, login)
	if err != nil {
		return nil, err
	}

	withdrawals, err := ordersStore.GetWithdrawals(ctx, login)
	if err != nil {
		return nil, err
	}

	for _, order := range processedOrders {
		accrual = order.Accrual.Add(accrual)
	}

	for _, withdraw := range withdrawals {
		withdrawn = withdraw.Sum.Add(withdrawn)
		accrual = accrual.Sub(withdraw.Sum)
	}

	balance = models.Balance{
		Current:   accrual,
		Withdrawn: withdrawn,
	}

	return &balance, nil
}
