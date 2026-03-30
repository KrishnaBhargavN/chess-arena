package game

type Game struct {
	ID      string
	PlayerA string
	PlayerB string
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
	game := Game{
		ID:      gameId,
		PlayerA: playerA,
		PlayerB: playerB,
	}
	gm.games[gameId] = &game
	gm.playerToGame[playerA] = &game
	gm.playerToGame[playerB] = &game

}
