package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/go-rfe/logging/log"
	"github.com/go-rfe/loyalty-system/internal/repository/orders"
)

const (
	pollTimeout     = 1 * time.Second
	accrualHTTPpath = "/api/orders/"
)

var (
	ErrOrderNotRegistered = errors.New("order doesn't registered")
	ErrTooManyRequests    = errors.New("wait for a while")
)

type PollerConfig struct {
	PollInterval   time.Duration
	AccrualAddress string
	AccrualScheme  string
}

type PollerWorker struct {
	Cfg PollerConfig
}

func (pw *PollerWorker) Run(ctx context.Context, ordersStore orders.Store) {
	pollTicker := time.NewTicker(pw.Cfg.PollInterval)
	defer pollTicker.Stop()

	client := http.Client{
		Timeout: pollTimeout,
	}

	serverURL := pw.Cfg.AccrualScheme + "://" + pw.Cfg.AccrualAddress + accrualHTTPpath

	storeContext, storeCancel := context.WithCancel(ctx)
	defer storeCancel()

	for {
		select {
		case <-ctx.Done():
			return
		case <-pollTicker.C:
			UpdateOrders(storeContext, client, serverURL, ordersStore)
		}
	}
}

func UpdateOrders(ctx context.Context, client http.Client, serverURL string, ordersStore orders.Store) {
	skipStatuses := map[string]struct{}{
		"REGISTERED": {},
	}

	getContext, getCancel := context.WithTimeout(ctx, pollTimeout)
	defer getCancel()

	ordersSlice, err := ordersStore.GetUnprocessedOrders(getContext)
	if err != nil {
		log.Error().Err(err).Msg("Poller couldn't get orders from store")
	}

	for _, order := range ordersSlice {
		orderGetURL := serverURL + order.Number
		accrualOrder, err := getOrder(getContext, orderGetURL, &client)
		if err != nil {
			log.Error().Err(err).Msg("filed to get order from accrual")

			continue
		}
		if _, ok := skipStatuses[accrualOrder.Status]; ok {
			continue
		}

		if err := ordersStore.UpdateOrder(ctx, accrualOrder); err != nil {
			log.Error().Err(err).Msg("filed to order order")
		}
	}
}

func getOrder(ctx context.Context, orderGetURL string, client *http.Client) (*orders.Order, error) {
	var order orders.Order

	accrualOrder := struct {
		Number  string  `json:"order"`
		Status  string  `json:"status"`
		Accrual float32 `json:"accrual,omitempty"`
	}{}

	log.Debug().Msgf("Update metric: %s", orderGetURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, orderGetURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Error().Err(err).Msg("couldn't close response body")
		}
	}(resp.Body)

	switch resp.StatusCode {
	case http.StatusOK:
		break
	case http.StatusNoContent:
		return nil, ErrOrderNotRegistered
	case http.StatusNotFound:
		return nil, ErrOrderNotRegistered
	case http.StatusTooManyRequests:
		return nil, ErrTooManyRequests
	default:
		return nil, fmt.Errorf("server response: %s", resp.Status)
	}

	err = json.NewDecoder(resp.Body).Decode(&accrualOrder)
	if err != nil {
		return nil, err
	}

	order = orders.Order{
		Number:  accrualOrder.Number,
		Status:  accrualOrder.Status,
		Accrual: accrualOrder.Accrual,
	}

	return &order, nil
}
