package accrual

import (
	"context"
	"errors"

	"github.com/shopspring/decimal"
)

var (
	ErrOrderNotRegistered = errors.New("order doesn't registered")
	ErrTooManyRequests    = errors.New("wait for a while")
)

type Client interface {
	GetOrder(ctx context.Context, orderID string) (*Accrual, error)
}

type Accrual struct {
	Number  string           `json:"order"`
	Status  string           `json:"status"`
	Accrual *decimal.Decimal `json:"Accrual,omitempty"`
}
