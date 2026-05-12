package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/notnil/chess"
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

	var payload struct {
		PlayerID string `json:"playerId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	res := a.Queue.JoinQueue(payload.PlayerID)

	if res == nil {
		resp := struct {
			Status string
		}{
			Status: "waiting",
		}
		writeJSON(w, http.StatusOK, resp)
		return
	}

	game, err := a.Store.CreateGame()
	if err != nil {
		http.Error(w, "failed to create game", http.StatusInternalServerError)
		return
	}

	a.Manager.AddGame(game.ID, res.PlayerA, res.PlayerB)
	msgA := ws.Message{
		Type:   "match_found",
		GameID: game.ID,
		Payload: json.RawMessage(
			fmt.Sprintf(`{
				"playerColor": "%s"
			}`, a.Manager.GetGame(game.ID).PlayerAColor),
		),
	}
	msgB := ws.Message{
		Type:   "match_found",
		GameID: game.ID,
		Payload: json.RawMessage(
			fmt.Sprintf(`{
				"playerColor": "%s"
			}`, a.Manager.GetGame(game.ID).PlayerBColor),
		),
	}
	if a.Hub != nil {
		_ = a.Hub.SendTo(res.PlayerA, msgA)
		_ = a.Hub.SendTo(res.PlayerB, msgB)
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"Status": "matched",
		"gameId": game.ID,
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
	}

	if err != nil {
		a.Logger.Println("get game: ", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, g)
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

	var payload struct {
		Move string `json:"move"`
		By   string `json:"by"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	pID := payload.By
	gg := a.Manager.GetGame(id)
	currentTurn := a.Store.GetTurn(id)
	var playerColor string
	if gg.PlayerA == pID {
		playerColor = gg.PlayerAColor
	} else if gg.PlayerB == pID {
		playerColor = gg.PlayerBColor
	} else {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if currentTurn == chess.Black {
		if playerColor == "white" {
			http.Error(w, "invalid move", http.StatusBadRequest)
			return
		}
	} else {
		if playerColor == "black" {
			http.Error(w, "invalid move", http.StatusBadRequest)
			return
		}
	}

	rec, err := a.Store.ApplyMove(id, payload.Move, payload.By)

	if err == store.ErrorNotFound {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	if err != nil {
		http.Error(w, "illegal move: "+err.Error(), http.StatusBadRequest)
		return
	}

	g, _ := a.Store.GetGame(id)

	resp := struct {
		Status string
		Move   models.MoveRecord
		FEN    string
	}{
		Status: "ok",
		Move:   rec,
		FEN:    g.CurrentFEN,
	}
	writeJSON(w, http.StatusOK, resp)

}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
