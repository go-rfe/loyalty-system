package users

import (
	"context"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

var ErrUserExists = errors.New("user already exists")
var ErrUserNotFound = errors.New("user not found")
var ErrInvalidLogin = errors.New("login is invalid")
var ErrInvalidPassword = errors.New("password is invalid")

type Store interface {
	CreateUser(ctx context.Context, login string, password string) error
	ValidateUser(ctx context.Context, login string, password string) error
}

type User struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func (u *User) Validate() error {
	if u.Login == "" {
		return ErrInvalidLogin
	}

	if u.Password == "" {
		return ErrInvalidPassword
	}

	return nil
}

func (u *User) CheckPassword(password string) error {
	return bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
}

func (u *User) SetPassword(password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	u.Password = string(hashedPassword)

	return nil
}
