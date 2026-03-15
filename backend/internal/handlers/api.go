package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"krishna.com/go-chess-backend/internal/models"
	"krishna.com/go-chess-backend/internal/store"
)

type API struct {
	Store store.Store
	Logger *log.Logger
}

func NewApi(s store.Store, l *log.Logger) *API {
	return &API{Store: s, Logger: l}
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
		By string `json:"by,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	rec, err := a.Store.ApplyMove(id, payload.Move, payload.By)

	if err == store.ErrorNotFound {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	if err != nil {
		http.Error(w, "illegal move: " + err.Error(), http.StatusBadRequest)
		return
	}

	g, _ := a.Store.GetGame(id)

	resp := struct {
		Status string
		Move models.MoveRecord
		FEN string
	} {
		Status: "ok",
		Move: rec,
		FEN: g.CurrentFEN,
	}
	writeJSON(w, http.StatusOK, resp)


}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}