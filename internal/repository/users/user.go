package users

import (
	"context"
	"errors"
)

var ErrUserExists = errors.New("user already exists")
var ErrUserNotFound = errors.New("user not found")

type Store interface {
	CreateUser(ctx context.Context, login string, password string) error
	ValidateUser(ctx context.Context, login string, password string) error
}
