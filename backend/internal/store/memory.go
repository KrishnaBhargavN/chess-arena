package store

import (
	"errors"
	"sync"

	"github.com/google/uuid"
	"krishna.com/go-chess-backend/internal/game"
	"krishna.com/go-chess-backend/internal/models"
)

var ErrorNotFound = errors.New("not found")

type Store interface {
	CreateGame() (models.Game, error)
	GetGame(id string) (models.Game, error)
	ApplyMove(gameID, moveStr, by string) (models.MoveRecord, error)
}

type MemoryStore struct {
	mu sync.RWMutex
	sessions map[string]*game.Session
	status map[string]string
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		sessions: make(map[string]*game.Session),
		status: make(map[string]string),
	}
}

func (s *MemoryStore) CreateGame() (models.Game, error) {
	id := uuid.New().String()
	sess := game.NewSession()
	s.mu.Lock()
	s.sessions[id] = sess
	s.status[id] = "waiting"
	s.mu.Unlock()

	return models.Game{
		ID: id,
		Status: "waiting",
		CurrentFEN: sess.FEN(),
		Moves: []models.MoveRecord{},
	}, nil
}

func (s *MemoryStore) GetGame(id string) (models.Game, error) {
	s.mu.RLock()
	sess, ok := s.sessions[id]
	st := s.status[id]
	s.mu.RUnlock()

	if !ok {
		return models.Game{}, ErrorNotFound
	}

	return models.Game{
		ID: id,
		Status: st,
		CurrentFEN: sess.FEN(),
		Moves: sess.Moves(),
	}, nil
}

func (s *MemoryStore) ApplyMove(gameID, moveStr, by string) (models.MoveRecord, error) {
	s.mu.RLock()
	sess, ok := s.sessions[gameID]
	s.mu.RUnlock()

	if !ok {
		return models.MoveRecord{}, ErrorNotFound
	}

	rec, err := sess.ApplyMove(moveStr, by)
	if err != nil {
		return models.MoveRecord{}, err
	}

	if rec.Ply == 0 {
		s.mu.Lock()
		s.status[gameID] = "playing"
		s.mu.Unlock()
	}
	
	return rec, nil
}