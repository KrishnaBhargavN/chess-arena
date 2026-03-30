package game

import (
	"fmt"
	"sync"
	"time"

	"github.com/notnil/chess"
	"krishna.com/go-chess-backend/internal/models"
)

type Session struct {
	mu         sync.Mutex
	game       *chess.Game
	initialFEN string
}

func NewSession() *Session {
	g := chess.NewGame()
	return &Session{
		game:       g,
		initialFEN: g.Position().String(),
	}
}

func (s *Session) FEN() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.game.Position().String()
}

func (s *Session) Moves() []models.MoveRecord {
	s.mu.Lock()
	defer s.mu.Unlock()

	out := make([]models.MoveRecord, 0, len(s.game.Moves()))
	positions := s.game.Positions()
	moves := s.game.Moves()

	for i, m := range moves {
		pos := positions[i]

		not := chess.AlgebraicNotation{}
		san := not.Encode(pos, m)
		rec := models.MoveRecord{
			Ply:       i,
			SAN:       san,
			UCI:       m.String(),
			From:      m.S1().String(),
			To:        m.S2().String(),
			Timestamp: time.Now(),
		}
		out = append(out, rec)
	}
	return out
}

func (s *Session) ApplyMove(moveStr string, by string) (models.MoveRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	beforePos := s.game.Position()
	fmt.Println("apply move")
	if err := s.game.MoveStr(moveStr); err != nil {
		fmt.Println("error with move")
		return models.MoveRecord{}, err
	}

	moves := s.game.Moves()
	last := moves[len(moves)-1]
	not := chess.AlgebraicNotation{}
	san := not.Encode(beforePos, last)

	rec := models.MoveRecord{
		Ply:       len(moves) - 1,
		SAN:       san,
		UCI:       last.String(),
		From:      last.S1().String(),
		To:        last.S2().String(),
		By:        by,
		Timestamp: time.Now(),
	}

	return rec, nil
}
