package accrual

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-rfe/logging/log"
	"github.com/go-rfe/loyalty-system/internal/models"
)

const (
	getTimeout      = 1 * time.Second
	accrualHTTPpath = "/api/orders/"
)

type client struct {
	httpClient http.Client
	serverURL  string
}

func NewAccrualClient(accrualSystemAddress string) *client {
	httpClient := http.Client{
		Timeout: getTimeout,
	}

	serverURL := accrualSystemAddress + accrualHTTPpath

	ac := client{
		httpClient: httpClient,
		serverURL:  serverURL,
	}

	return &ac
}

func (c *client) GetOrder(ctx context.Context, orderID string) (*models.Order, error) {
	order := models.Order{
		Number: orderID,
	}
	accrualOrder := accrual{}

	orderGetURL := c.serverURL + order.Number
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, orderGetURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

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
	err = resp.Body.Close()
	if err != nil {
		log.Error().Err(err).Msg("couldn't close response body")
	}

	order.Status = accrualOrder.Status
	order.Accrual = accrualOrder.Accrual

	return &order, nil
}
