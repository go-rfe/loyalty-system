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
	"github.com/go-rfe/loyalty-system/internal/models"
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

		if err := models.Validate(string(orderNumber)); err != nil {
			http.Error(
				w,
				fmt.Sprintf("Bad order number: %s (%q)", orderNumber, err),
				http.StatusUnprocessableEntity,
			)

			return
		}

		login, err := getLoginFromRequest(r)
		if err != nil {
			http.Error(
				w,
				fmt.Sprintf("couldn't get user from token: %q", err),
				http.StatusInternalServerError,
			)

			return
		}

		err = ordersStore.CreateOrder(requestContext, login, string(orderNumber))
		switch {
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

		login, err := getLoginFromRequest(r)
		if err != nil {
			http.Error(
				w,
				fmt.Sprintf("couldn't get user from token: %q", err),
				http.StatusInternalServerError,
			)

			return
		}

		ordersSlice, err := ordersStore.GetOrders(requestContext, login)
		if err != nil {
			log.Error().Err(err).Msgf("couldn't get orders for %s", login)
			http.Error(
				w,
				fmt.Sprintf("couldn't get orders for %s: %q", login, err),
				http.StatusInternalServerError,
			)

			return
		}

		w.Header().Set("Content-Type", "application/json")
		err = models.Encode(&ordersSlice, w)
		if err != nil {
			log.Error().Err(err).Msg("Cannot send request")
		}
	}
}

func getLoginFromRequest(r *http.Request) (string, error) {
	_, claims, err := jwtauth.FromContext(r.Context())
	if err != nil {
		return "", ErrInvalidToken
	}

	login, ok := claims["sub"].(string)
	if !ok {
		return "", ErrInvalidToken
	}

	return login, nil
}
