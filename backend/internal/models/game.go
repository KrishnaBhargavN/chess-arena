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
	ID              string       `json:"id"`
	PlayerA         string       `json:"playerA,omitempty"`
	PlayerB         string       `json:"playerB,omitempty"`
	PlayerAUsername string       `json:"playerAUsername,omitempty"`
	PlayerBUsername string       `json:"playerBUsername,omitempty"`
	PlayerAColor    string       `json:"playerAColor,omitempty"`
	PlayerBColor    string       `json:"playerBColor,omitempty"`
	Status          string       `json:"status"`
	CurrentFEN      string       `json:"initial_fen"`
	Moves           []MoveRecord `json:"moves"`
	CreatedAt       time.Time    `json:"createdAt,omitempty"`
	FinishedAt      *time.Time   `json:"finishedAt,omitempty"`
}
