package handlers

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"

	"github.com/go-rfe/logging/log"
	"github.com/go-rfe/loyalty-system/internal/repository/orders"
)

func OrdersHandler(ordersStore orders.Store) func(r chi.Router) {
	return func(r chi.Router) {
		r.Post("/", createOrder(ordersStore))
		r.Get("/", getOrders(ordersStore))
	}
}

func createOrder(ordersStore orders.Store) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		requestContext, requestCancel := context.WithTimeout(r.Context(), requestTimeout)
		defer requestCancel()

		orderNumber, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(
				w,
				fmt.Sprintf("Could'n read order number: %v", err),
				http.StatusBadRequest,
			)

			return
		}

		if err := orders.Validate(string(orderNumber)); err != nil {
			http.Error(
				w,
				fmt.Sprintf("Bad order number: %s (%q)", orderNumber, err),
				http.StatusUnprocessableEntity,
			)

			return
		}

		_, claims, err := jwtauth.FromContext(r.Context())
		if err != nil {
			http.Error(
				w,
				fmt.Sprintf("couldn't get user from token: %q", err),
				http.StatusInternalServerError,
			)

			return
		}

		login := fmt.Sprintf("%v", claims["sub"])
		switch err := ordersStore.CreateOrder(requestContext, login, string(orderNumber)); {
		case errors.Is(err, orders.ErrOtherOrderExists):
			http.Error(
				w,
				err.Error(),
				http.StatusConflict,
			)
		case errors.Is(err, orders.ErrOrderExists):
			w.WriteHeader(http.StatusOK)
		case err != nil:
			http.Error(
				w,
				fmt.Sprintf("couldn't create order: %q", err),
				http.StatusInternalServerError,
			)
		default:
			w.WriteHeader(http.StatusAccepted)
		}
	}
}

func getOrders(ordersStore orders.Store) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		requestContext, requestCancel := context.WithTimeout(r.Context(), requestTimeout)
		defer requestCancel()

		_, claims, err := jwtauth.FromContext(r.Context())
		if err != nil {
			http.Error(
				w,
				fmt.Sprintf("couldn't get user from token: %q", err),
				http.StatusInternalServerError,
			)

			return
		}

		login := fmt.Sprintf("%v", claims["sub"])

		ordersSlice, err := ordersStore.GetOrders(requestContext, login)
		if err != nil {
			http.Error(
				w,
				fmt.Sprintf("couldn't get orders for %s: %q", login, err),
				http.StatusInternalServerError,
			)

			return
		}

		encodedOrders, err := orders.Encode(&ordersSlice)
		if err != nil {
			http.Error(
				w,
				fmt.Sprintf("couldn't encode orders: %q", err),
				http.StatusInternalServerError,
			)

			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write(encodedOrders.Bytes())
		if err != nil {
			log.Error().Err(err).Msg("Cannot send request")
		}
	}
}
