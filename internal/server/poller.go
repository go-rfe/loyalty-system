package server

import (
	"context"
	"time"

	"github.com/go-rfe/logging/log"
	"github.com/go-rfe/loyalty-system/internal/accrual"
	"github.com/go-rfe/loyalty-system/internal/repository/orders"
)

const (
	pollTimeout = 1 * time.Second
)

type PollerConfig struct {
	PollInterval   time.Duration
	AccrualAddress string
}

type PollerWorker struct {
	Cfg PollerConfig
}

func (pw *PollerWorker) Run(ctx context.Context, ordersStore orders.Store) {
	pollTicker := time.NewTicker(pw.Cfg.PollInterval)
	defer pollTicker.Stop()

	accrualClient := accrual.NewAccrualClient(pw.Cfg.AccrualAddress)

	for {
		select {
		case <-ctx.Done():
			return
		case <-pollTicker.C:
			UpdateOrders(ctx, accrualClient, ordersStore)
		}
	}
}

func UpdateOrders(ctx context.Context, accrualClient accrual.Client, ordersStore orders.Store) {
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
		accrualOrder, err := accrualClient.GetOrder(getContext, order.Number)
		if err != nil {
			log.Error().Err(err).Msg("filed to get order from accrual")

			continue
		}
		if _, ok := skipStatuses[accrualOrder.Status]; ok {
			continue
		}

		if err := ordersStore.UpdateOrder(ctx, accrualOrder); err != nil {
			log.Error().Err(err).Msgf("filed to update %s order", accrualOrder.Number)
		}
	}
}
