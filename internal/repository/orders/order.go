package orders

import (
	"context"
	"errors"

	"github.com/go-rfe/loyalty-system/internal/models"
)

var (
	ErrOrderExists      = errors.New("order already exists")
	ErrOtherOrderExists = errors.New("other user order already exists")
)

type Store interface {
	CreateOrder(ctx context.Context, login string, order string) error
	UpdateOrder(ctx context.Context, order *models.Order) error
	GetOrders(ctx context.Context, login string) ([]models.Order, error)
	GetUnprocessedOrders(ctx context.Context) ([]models.Order, error)
	GetProcessedOrders(ctx context.Context, login string) ([]models.Order, error)
	Withdraw(ctx context.Context, login string, withdraw *models.Withdraw) error
	GetWithdrawals(ctx context.Context, login string) ([]models.Withdraw, error)
}
