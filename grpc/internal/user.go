package internal

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

var (
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrUserNotFound      = errors.New("user not found")
)

type (
	User struct {
		ID        string
		Name      string
		Surname   string
		Age       int
		CreatedAt time.Time
		UpdatedAt time.Time
		Disabled  bool
	}

	UserService struct {
		store map[string]User
		mx    *sync.RWMutex
	}
)

func NewUserService() *UserService {
	return &UserService{
		store: make(map[string]User),
		mx:    &sync.RWMutex{},
	}
}

func (s *UserService) Create(user User) (*User, error) {
	s.mx.RLock()

	for _, u := range s.store {
		if u.Name == user.Name {
			s.mx.RUnlock()
			return nil, fmt.Errorf("%w: %s", ErrUserAlreadyExists, u.Name)
		}
	}
	s.mx.RUnlock()

	s.mx.Lock()
	defer s.mx.Unlock()
	user.ID = uuid.New().String()
	s.store[user.ID] = user

	return &user, nil
}

func (s *UserService) Update(user User) (*User, error) {
	s.mx.Lock()
	defer s.mx.Unlock()

	_, ok := s.store[user.ID]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrUserNotFound, user.ID)
	}

	user.UpdatedAt = time.Now()
	s.store[user.ID] = user

	return &user, nil
}

func (s *UserService) Get(id string) (*User, error) {
	s.mx.RLock()
	defer s.mx.RUnlock()

	user, ok := s.store[id]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrUserNotFound, id)
	}

	return &user, nil
}

func (s *UserService) Delete(id string) error {
	s.mx.Lock()
	defer s.mx.Unlock()

	_, ok := s.store[id]
	if !ok {
		return fmt.Errorf("%w: %s", ErrUserNotFound, id)
	}

	delete(s.store, id)

	return nil
}
