package accrual

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
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

func (c *client) GetOrder(ctx context.Context, orderID string) (*Accrual, error) {
	accrualOrder := Accrual{}

	orderGetURL := c.serverURL + orderID
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, orderGetURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	err = checkStatusCode(resp)
	if err != nil {
		return nil, err
	}

	err = json.NewDecoder(resp.Body).Decode(&accrualOrder)
	if err != nil {
		return nil, err
	}

	return &accrualOrder, nil
}

func checkStatusCode(resp *http.Response) error {
	switch resp.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusNoContent:
		return ErrOrderNotRegistered
	case http.StatusNotFound:
		return ErrOrderNotRegistered
	case http.StatusTooManyRequests:
		return ErrTooManyRequests
	default:
		return fmt.Errorf("server response: %s", resp.Status)
	}
}
