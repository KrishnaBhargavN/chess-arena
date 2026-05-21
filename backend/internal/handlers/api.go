package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/notnil/chess"
	"krishna.com/go-chess-backend/internal/auth"
	"krishna.com/go-chess-backend/internal/game"
	"krishna.com/go-chess-backend/internal/matchmaking"
	"krishna.com/go-chess-backend/internal/models"
	"krishna.com/go-chess-backend/internal/store"
	"krishna.com/go-chess-backend/internal/ws"
)

type API struct {
	Store   store.Store
	Logger  *log.Logger
	Queue   *matchmaking.Queue
	Hub     *ws.Hub
	Manager *game.GameManager
}

func NewApi(s store.Store, l *log.Logger, q *matchmaking.Queue, hub *ws.Hub, manager *game.GameManager) *API {
	return &API{Store: s, Logger: l, Queue: q, Hub: hub, Manager: manager}
}

func (a *API) CreateGame(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	g, err := a.Store.CreateGame()
	if err != nil {
		a.Logger.Println("create game: ", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, g)
}

func (a *API) JoinMatchMaking(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	playerID := auth.UserIDFromContext(r.Context())
	if playerID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	res := a.Queue.JoinQueue(playerID)

	if res == nil {
		writeJSON(w, http.StatusOK, struct{ Status string }{Status: "waiting"})
		return
	}

	g, err := a.Store.CreateGame()
	if err != nil {
		http.Error(w, "failed to create game", http.StatusInternalServerError)
		return
	}

	a.Manager.AddGame(g.ID, res.PlayerA, res.PlayerB)
	gm := a.Manager.GetGame(g.ID)
	_ = a.Store.UpdateGamePlayers(g.ID, res.PlayerA, res.PlayerB, gm.PlayerAColor, gm.PlayerBColor)

	msgA := ws.Message{
		Type:   "match_found",
		GameID: g.ID,
		Payload: json.RawMessage(
			fmt.Sprintf(`{"playerColor":"%s"}`, gm.PlayerAColor),
		),
	}
	msgB := ws.Message{
		Type:   "match_found",
		GameID: g.ID,
		Payload: json.RawMessage(
			fmt.Sprintf(`{"playerColor":"%s"}`, gm.PlayerBColor),
		),
	}
	if a.Hub != nil {
		_ = a.Hub.SendToLobby(res.PlayerA, msgA)
		_ = a.Hub.SendToLobby(res.PlayerB, msgB)
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"Status":      "matched",
		"gameId":      g.ID,
		"playerColor": gm.PlayerBColor,
	})
}

func (a *API) GetGame(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/games/")
	if id == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}

	g, err := a.Store.GetGame(id)
	if err == store.ErrorNotFound {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if err != nil {
		a.Logger.Println("get game: ", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, g)
}

func (a *API) ListGames(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	games, err := a.Store.ListGames(userID)
	if err != nil {
		a.Logger.Println("list games: ", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, games)
}

func (a *API) GetMoves(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/games/"), "/")
	if len(parts) < 2 || parts[1] != "moves" {
		http.Error(w, "bad path", http.StatusBadRequest)
		return
	}
	id := parts[0]

	g, err := a.Store.GetGame(id)
	if err == store.ErrorNotFound {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if err != nil {
		a.Logger.Println("get moves: ", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, g.Moves)
}

func (a *API) MakeMove(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/games/"), "/")
	if len(parts) < 2 || parts[1] != "move" {
		http.Error(w, "bad path", http.StatusBadRequest)
		return
	}
	id := parts[0]

	playerID := auth.UserIDFromContext(r.Context())
	if playerID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var payload struct {
		Move string `json:"move"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	gg := a.Manager.GetGame(id)
	if gg == nil {
		http.Error(w, "game not found", http.StatusNotFound)
		return
	}

	currentTurn := a.Store.GetTurn(id)
	var playerColor string
	if gg.PlayerA == playerID {
		playerColor = gg.PlayerAColor
	} else if gg.PlayerB == playerID {
		playerColor = gg.PlayerBColor
	} else {
		http.Error(w, "not a participant", http.StatusForbidden)
		return
	}

	if currentTurn == chess.Black {
		if playerColor == "white" {
			http.Error(w, "not your turn", http.StatusBadRequest)
			return
		}
	} else {
		if playerColor == "black" {
			http.Error(w, "not your turn", http.StatusBadRequest)
			return
		}
	}

	rec, err := a.Store.ApplyMove(id, payload.Move, playerID)
	if err == store.ErrorNotFound {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "illegal move: "+err.Error(), http.StatusBadRequest)
		return
	}

	g, _ := a.Store.GetGame(id)
	writeJSON(w, http.StatusOK, struct {
		Status string
		Move   models.MoveRecord
		FEN    string
	}{Status: "ok", Move: rec, FEN: g.CurrentFEN})
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
