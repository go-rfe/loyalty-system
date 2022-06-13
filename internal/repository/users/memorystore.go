package users

import (
	"context"
	"sync"

	"github.com/go-rfe/loyalty-system/internal/models"
)

type InMemoryStore struct {
	usersCache map[string]*models.User
	mu         sync.RWMutex
}

func NewInMemoryStore() *InMemoryStore {
	m := InMemoryStore{
		usersCache: make(map[string]*models.User),
	}

	return &m
}

func (m *InMemoryStore) CreateUser(_ context.Context, login string, password string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.usersCache[login]; ok {
		return ErrUserExists
	}

	user := &models.User{
		Login:    login,
		Password: password,
	}

	m.usersCache[login] = user

	return nil
}

func (m *InMemoryStore) GetUser(_ context.Context, login string) (*models.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	user, ok := m.usersCache[login]
	if !ok {
		return nil, ErrUserNotFound
	}

	return user, nil
}
