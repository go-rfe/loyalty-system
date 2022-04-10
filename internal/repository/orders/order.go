package orders

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/go-rfe/utils/luhn"
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
	GetBalance(ctx context.Context, login string) (*Balance, error)
	Withdraw(ctx context.Context, login string, withdraw *Withdraw) error
	GetWithdrawals(ctx context.Context, login string) ([]Withdraw, error)
}

type Order struct {
	Number     string    `json:"number"`
	Status     string    `json:"status"`
	Accrual    float32   `json:"accrual,omitempty"`
	UploadedAt time.Time `json:"uploaded_at"`
}

type Balance struct {
	Current   float32 `json:"current"`
	Withdrawn float32 `json:"withdrawn"`
}

type Withdraw struct {
	Order       string    `json:"order"`
	Sum         float32   `json:"sum"`
	ProcessedAt time.Time `json:"processed_at"`
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

func Encode(data interface{}) (*bytes.Buffer, error) {
	var buf bytes.Buffer
	jsonEncoder := json.NewEncoder(&buf)

	if err := jsonEncoder.Encode(data); err != nil {
		return nil, err
	}

	return &buf, nil
}
