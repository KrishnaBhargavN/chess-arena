package game

import "math/rand"

type Game struct {
	ID           string
	PlayerA      string
	PlayerB      string
	PlayerAColor string
	PlayerBColor string
}

type GameManager struct {
	games        map[string]*Game
	playerToGame map[string]*Game
}

func NewGameManager() *GameManager {
	return &GameManager{
		games:        make(map[string]*Game),
		playerToGame: make(map[string]*Game),
	}
}

func (gm *GameManager) GetGame(id string) *Game {
	return gm.games[id]
}

func (gm *GameManager) AddGame(gameId, playerA, playerB string) {
	playerAColor := "white"
	playerBColor := "black"
	if rand.Intn(2) == 0 {
		playerAColor = "black"
		playerBColor = "white"
	}
	game := Game{
		ID:           gameId,
		PlayerA:      playerA,
		PlayerB:      playerB,
		PlayerAColor: playerAColor,
		PlayerBColor: playerBColor,
	}
	gm.games[gameId] = &game
	gm.playerToGame[playerA] = &game
	gm.playerToGame[playerB] = &game

}
