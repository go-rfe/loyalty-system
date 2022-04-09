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
	ErrOrderExists        = errors.New("order already exists")
	ErrOtherOrderExists   = errors.New("other user order already exists")
	ErrInvalidOrderNumber = errors.New("order number is invalid")
)

type Store interface {
	CreateOrder(ctx context.Context, login string, order *Order) error
	GetOrders(ctx context.Context, login string) ([]Order, error)
}

type Order struct {
	Number     string    `json:"number"`
	Status     string    `json:"status"`
	Accrual    float32   `json:"accrual"`
	UploadedAt time.Time `json:"uploaded_at"`
}

func (o *Order) Validate() error {
	numberInt, err := strconv.ParseInt(o.Number, orderNumberBase, orderNumberBitSize)
	if err != nil {
		return err
	}

	if !luhn.Valid(numberInt) {
		return ErrInvalidOrderNumber
	}

	return nil
}

func Encode(orders *[]Order) (*bytes.Buffer, error) {
	var buf bytes.Buffer
	jsonEncoder := json.NewEncoder(&buf)

	if err := jsonEncoder.Encode(orders); err != nil {
		return nil, err
	}

	return &buf, nil
}
