package auth

import (
	"errors"
	"sync"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var ErrUserExists = errors.New("user already exists")
var ErrInvalidCredentials = errors.New("invalid credentials")
var ErrUserNotFound = errors.New("user not found")

type UserStorer interface {
	Create(username, password string) (*User, error)
	Authenticate(username, password string) (*User, error)
}

type User struct {
	ID           string
	Username     string
	PasswordHash string
}

type UserStore struct {
	mu    sync.RWMutex
	byUsername map[string]*User
	byID       map[string]*User
}

func NewUserStore() *UserStore {
	return &UserStore{
		byUsername: make(map[string]*User),
		byID:       make(map[string]*User),
	}
}

func (s *UserStore) Create(username, password string) (*User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.byUsername[username]; exists {
		return nil, ErrUserExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	u := &User{
		ID:           uuid.NewString(),
		Username:     username,
		PasswordHash: string(hash),
	}
	s.byUsername[username] = u
	s.byID[u.ID] = u
	return u, nil
}

func (s *UserStore) Authenticate(username, password string) (*User, error) {
	s.mu.RLock()
	u, exists := s.byUsername[username]
	s.mu.RUnlock()

	if !exists {
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	return u, nil
}
