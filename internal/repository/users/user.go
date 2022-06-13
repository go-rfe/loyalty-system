package users

import (
	"context"
	"errors"

	"github.com/go-rfe/loyalty-system/internal/models"
)

var ErrUserExists = errors.New("user already exists")
var ErrUserNotFound = errors.New("user not found")

type Store interface {
	CreateUser(ctx context.Context, login string, password string) error
	GetUser(ctx context.Context, login string) (*models.User, error)
}
