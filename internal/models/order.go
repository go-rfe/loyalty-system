package models

import (
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

var ErrInvalidOrderNumber = errors.New("order number is invalid")

type Order struct {
	Number     string           `json:"number"`
	Status     string           `json:"status"`
	Accrual    *decimal.Decimal `json:"accrual,omitempty"`
	UploadedAt time.Time        `json:"uploaded_at"`
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
