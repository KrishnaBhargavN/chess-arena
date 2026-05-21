package store

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/notnil/chess"
	"krishna.com/go-chess-backend/internal/game"
	"krishna.com/go-chess-backend/internal/models"
)

type PostgresStore struct {
	db       *pgxpool.Pool
	mu       sync.RWMutex
	sessions map[string]*game.Session
}

func NewPostgresStore(db *pgxpool.Pool) *PostgresStore {
	return &PostgresStore{
		db:       db,
		sessions: make(map[string]*game.Session),
	}
}

func (s *PostgresStore) getOrLoadSession(gameID string) (*game.Session, error) {
	s.mu.RLock()
	sess, ok := s.sessions[gameID]
	s.mu.RUnlock()
	if ok {
		return sess, nil
	}

	var exists bool
	if err := s.db.QueryRow(context.Background(),
		`SELECT EXISTS(SELECT 1 FROM games WHERE id=$1)`, gameID,
	).Scan(&exists); err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrorNotFound
	}

	rows, err := s.db.Query(context.Background(),
		`SELECT san FROM moves WHERE game_id=$1 ORDER BY ply`, gameID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rebuilt := game.NewSession()
	for rows.Next() {
		var san string
		if err := rows.Scan(&san); err != nil {
			return nil, err
		}
		if _, err := rebuilt.ApplyMove(san, ""); err != nil {
			return nil, err
		}
	}

	s.mu.Lock()
	if existing, ok := s.sessions[gameID]; ok {
		rebuilt = existing
	} else {
		s.sessions[gameID] = rebuilt
	}
	s.mu.Unlock()
	return rebuilt, nil
}

func (s *PostgresStore) CreateGame() (models.Game, error) {
	id := uuid.NewString()
	sess := game.NewSession()
	fen := sess.FEN()

	_, err := s.db.Exec(context.Background(),
		`INSERT INTO games(id, status, current_fen) VALUES($1, 'waiting', $2)`,
		id, fen,
	)
	if err != nil {
		return models.Game{}, err
	}

	s.mu.Lock()
	s.sessions[id] = sess
	s.mu.Unlock()

	return models.Game{
		ID:         id,
		Status:     "waiting",
		CurrentFEN: fen,
		Moves:      []models.MoveRecord{},
	}, nil
}

func (s *PostgresStore) GetGame(id string) (models.Game, error) {
	var g models.Game
	var playerA, playerB, usernameA, usernameB, colorA, colorB *string
	var finishedAt *time.Time
	err := s.db.QueryRow(context.Background(),
		`SELECT g.player_a_id::text, g.player_b_id::text, ua.username, ub.username,
		        g.player_a_color, g.player_b_color,
		        g.status, g.current_fen, g.created_at, g.finished_at
		 FROM games g
		 LEFT JOIN users ua ON g.player_a_id = ua.id
		 LEFT JOIN users ub ON g.player_b_id = ub.id
		 WHERE g.id=$1`, id,
	).Scan(&playerA, &playerB, &usernameA, &usernameB, &colorA, &colorB,
		&g.Status, &g.CurrentFEN, &g.CreatedAt, &finishedAt)
	if err != nil {
		return models.Game{}, ErrorNotFound
	}
	g.ID = id
	if playerA != nil {
		g.PlayerA = *playerA
	}
	if playerB != nil {
		g.PlayerB = *playerB
	}
	if usernameA != nil {
		g.PlayerAUsername = *usernameA
	}
	if usernameB != nil {
		g.PlayerBUsername = *usernameB
	}
	if colorA != nil {
		g.PlayerAColor = *colorA
	}
	if colorB != nil {
		g.PlayerBColor = *colorB
	}
	g.FinishedAt = finishedAt

	rows, err := s.db.Query(context.Background(),
		`SELECT ply, san, uci, from_sq, to_sq, COALESCE(played_by::text, ''), played_at
		 FROM moves WHERE game_id=$1 ORDER BY ply`, id,
	)
	if err != nil {
		return models.Game{}, err
	}
	defer rows.Close()

	moves := []models.MoveRecord{}
	for rows.Next() {
		var rec models.MoveRecord
		if err := rows.Scan(&rec.Ply, &rec.SAN, &rec.UCI, &rec.From, &rec.To, &rec.By, &rec.Timestamp); err != nil {
			return models.Game{}, err
		}
		moves = append(moves, rec)
	}
	g.Moves = moves
	return g, nil
}

func (s *PostgresStore) ListGames(userID string) ([]models.Game, error) {
	rows, err := s.db.Query(context.Background(),
		`SELECT g.id::text, g.player_a_id::text, g.player_b_id::text,
		        ua.username, ub.username,
		        g.player_a_color, g.player_b_color,
		        g.status, g.current_fen, g.created_at, g.finished_at
		 FROM games g
		 LEFT JOIN users ua ON g.player_a_id = ua.id
		 LEFT JOIN users ub ON g.player_b_id = ub.id
		 WHERE g.player_a_id=$1 OR g.player_b_id=$1
		 ORDER BY g.created_at DESC`, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []models.Game{}
	for rows.Next() {
		var g models.Game
		var playerA, playerB, usernameA, usernameB, colorA, colorB *string
		var finishedAt *time.Time
		if err := rows.Scan(&g.ID, &playerA, &playerB, &usernameA, &usernameB,
			&colorA, &colorB, &g.Status, &g.CurrentFEN, &g.CreatedAt, &finishedAt); err != nil {
			return nil, err
		}
		if playerA != nil {
			g.PlayerA = *playerA
		}
		if playerB != nil {
			g.PlayerB = *playerB
		}
		if usernameA != nil {
			g.PlayerAUsername = *usernameA
		}
		if usernameB != nil {
			g.PlayerBUsername = *usernameB
		}
		if colorA != nil {
			g.PlayerAColor = *colorA
		}
		if colorB != nil {
			g.PlayerBColor = *colorB
		}
		g.FinishedAt = finishedAt
		out = append(out, g)
	}
	return out, nil
}

func (s *PostgresStore) ApplyMove(gameID, moveStr, by string) (models.MoveRecord, error) {
	sess, err := s.getOrLoadSession(gameID)
	if err != nil {
		return models.MoveRecord{}, err
	}

	rec, err := sess.ApplyMove(moveStr, by)
	if err != nil {
		return models.MoveRecord{}, err
	}

	newFEN := sess.FEN()
	status := "playing"

	if sess.Outcome() != "" {
		status = "finished"
	}

	_, err = s.db.Exec(context.Background(),
		`UPDATE games SET status=$1, current_fen=$2, finished_at=CASE WHEN $1='finished' THEN NOW() ELSE NULL END WHERE id=$3`,
		status, newFEN, gameID,
	)
	if err != nil {
		return models.MoveRecord{}, err
	}

	_, err = s.db.Exec(context.Background(),
		`INSERT INTO moves(game_id, ply, san, uci, from_sq, to_sq, played_by, played_at)
		 VALUES($1, $2, $3, $4, $5, $6, $7::uuid, $8)`,
		gameID, rec.Ply, rec.SAN, rec.UCI, rec.From, rec.To, by, rec.Timestamp,
	)
	if err != nil {
		return models.MoveRecord{}, err
	}

	return rec, nil
}

func (s *PostgresStore) GetTurn(gameID string) chess.Color {
	sess, err := s.getOrLoadSession(gameID)
	if err != nil || sess == nil {
		return chess.White
	}
	return sess.Turn()
}

func (s *PostgresStore) UpdateGamePlayers(gameID, playerA, playerB, colorA, colorB string) error {
	_, err := s.db.Exec(context.Background(),
		`UPDATE games SET player_a_id=$1, player_b_id=$2, player_a_color=$3, player_b_color=$4 WHERE id=$5`,
		playerA, playerB, colorA, colorB, gameID,
	)
	return err
}

func (s *PostgresStore) Outcome(gameID string) string {
	sess, err := s.getOrLoadSession(gameID)
	if err != nil || sess == nil {
		return ""
	}
	return sess.Outcome()
}
