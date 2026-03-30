package models

import "time"

type MoveRecord struct {
	Ply       int       `json:"ply"`
	SAN       string    `json:"san"`
	UCI       string    `json:"uci"`
	From      string    `json:"from"`
	To        string    `json:"to"`
	By        string    `json:"by,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

type Game struct {
	ID         string       `json:"id"`
	Status     string       `json:"status"`
	CurrentFEN string       `json:"initial_fen"`
	Moves      []MoveRecord `json:"moves"`
}
