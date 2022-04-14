package orders

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"strconv"
	"time"

	"github.com/go-rfe/utils/luhn"
	"github.com/shopspring/decimal"
)

const (
	orderNumberBase    = 10
	orderNumberBitSize = 64
)

var (
	ErrOrderExists         = errors.New("order already exists")
	ErrOtherOrderExists    = errors.New("other user order already exists")
	ErrInvalidOrderNumber  = errors.New("order number is invalid")
	ErrInsufficientBalance = errors.New("insufficient balance")
)

type Store interface {
	CreateOrder(ctx context.Context, login string, order string) error
	UpdateOrder(ctx context.Context, order *Order) error
	GetOrders(ctx context.Context, login string) ([]Order, error)
	GetUnprocessedOrders(ctx context.Context) ([]Order, error)
	GetProcessedOrders(ctx context.Context, login string) ([]Order, error)
	Withdraw(ctx context.Context, login string, withdraw *Withdraw) error
	GetWithdrawals(ctx context.Context, login string) ([]Withdraw, error)
}

type Order struct {
	Number     string          `json:"number"`
	Status     string          `json:"status"`
	Accrual    decimal.Decimal `json:"accrual,omitempty"`
	UploadedAt time.Time       `json:"uploaded_at"`
}

type Balance struct {
	Current   decimal.Decimal `json:"current"`
	Withdrawn decimal.Decimal `json:"withdrawn"`
}

type Withdraw struct {
	Order       string          `json:"order"`
	Sum         decimal.Decimal `json:"sum"`
	ProcessedAt time.Time       `json:"processed_at"`
}

func Validate(number string) error {
	numberInt, err := strconv.ParseInt(number, orderNumberBase, orderNumberBitSize)
	if err != nil {
		return err
	}

	if !luhn.Valid(numberInt) {
		return ErrInvalidOrderNumber
	}

	return nil
}

func Encode(data interface{}, w io.Writer) error {
	jsonEncoder := json.NewEncoder(w)
	decimal.MarshalJSONWithoutQuotes = true
	return jsonEncoder.Encode(data)
}

func GetBalance(ctx context.Context, login string, ordersStore Store) (*Balance, error) {
	var (
		balance   Balance
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

	balance = Balance{
		Current:   accrual,
		Withdrawn: withdrawn,
	}

	return &balance, nil
}
