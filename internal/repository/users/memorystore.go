package users

import (
	"context"
	"sync"
)

type InMemoryStore struct {
	usersCache map[string]*User
	mu         sync.RWMutex
}

func NewInMemoryStore() *InMemoryStore {
	m := InMemoryStore{
		usersCache: make(map[string]*User),
	}

	return &m
}

func (m *InMemoryStore) CreateUser(_ context.Context, login string, password string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.usersCache[login]; ok {
		return ErrUserExists
	}

	user := &User{
		Login: login,
	}

	if err := user.SetPassword(password); err != nil {
		return err
	}

	m.usersCache[login] = user

	return nil
}

func (m *InMemoryStore) ValidateUser(_ context.Context, login string, password string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	user, ok := m.usersCache[login]
	if !ok {
		return ErrUserNotFound
	}

	if err := user.CheckPassword(password); err != nil {
		return ErrInvalidPassword
	}

	return nil
}
