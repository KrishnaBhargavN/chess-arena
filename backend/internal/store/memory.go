package store

import (
	"errors"
	"sync"

	"github.com/google/uuid"
	"github.com/notnil/chess"
	"krishna.com/go-chess-backend/internal/game"
	"krishna.com/go-chess-backend/internal/models"
)

var ErrorNotFound = errors.New("not found")

type Store interface {
	CreateGame() (models.Game, error)
	GetGame(id string) (models.Game, error)
	ListGames(userID string) ([]models.Game, error)
	ApplyMove(gameID, moveStr, by string) (models.MoveRecord, error)
	GetTurn(gameID string) chess.Color
	UpdateGamePlayers(gameID, playerA, playerB, colorA, colorB string) error
	Outcome(gameID string) string
}

type playerInfo struct {
	playerA, playerB, colorA, colorB string
}

type MemoryStore struct {
	mu       sync.RWMutex
	sessions map[string]*game.Session
	status   map[string]string
	players  map[string]playerInfo
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		sessions: make(map[string]*game.Session),
		status:   make(map[string]string),
		players:  make(map[string]playerInfo),
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
		ID:         id,
		Status:     "waiting",
		CurrentFEN: sess.FEN(),
		Moves:      []models.MoveRecord{},
	}, nil
}

func (s *MemoryStore) GetGame(id string) (models.Game, error) {
	s.mu.RLock()
	sess, ok := s.sessions[id]
	st := s.status[id]
	p := s.players[id]
	s.mu.RUnlock()

	if !ok {
		return models.Game{}, ErrorNotFound
	}

	return models.Game{
		ID:           id,
		PlayerA:      p.playerA,
		PlayerB:      p.playerB,
		PlayerAColor: p.colorA,
		PlayerBColor: p.colorB,
		Status:       st,
		CurrentFEN:   sess.FEN(),
		Moves:        sess.Moves(),
	}, nil
}

func (s *MemoryStore) ListGames(userID string) ([]models.Game, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := []models.Game{}
	for id, p := range s.players {
		if p.playerA != userID && p.playerB != userID {
			continue
		}
		sess := s.sessions[id]
		if sess == nil {
			continue
		}
		out = append(out, models.Game{
			ID:           id,
			PlayerA:      p.playerA,
			PlayerB:      p.playerB,
			PlayerAColor: p.colorA,
			PlayerBColor: p.colorB,
			Status:       s.status[id],
			CurrentFEN:   sess.FEN(),
		})
	}
	return out, nil
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

	s.mu.Lock()
	if sess.Outcome() != "" {
		s.status[gameID] = "finished"
	} else {
		s.status[gameID] = "playing"
	}
	s.mu.Unlock()

	return rec, nil
}

func (s *MemoryStore) GetTurn(gameID string) chess.Color {
	sess := s.sessions[gameID]
	return sess.Turn()
}

func (s *MemoryStore) UpdateGamePlayers(gameID, playerA, playerB, colorA, colorB string) error {
	s.mu.Lock()
	s.players[gameID] = playerInfo{playerA, playerB, colorA, colorB}
	s.mu.Unlock()
	return nil
}

func (s *MemoryStore) Outcome(gameID string) string {
	s.mu.RLock()
	sess := s.sessions[gameID]
	s.mu.RUnlock()
	if sess == nil {
		return ""
	}
	return sess.Outcome()
}
