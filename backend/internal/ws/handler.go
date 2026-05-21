package ws

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"krishna.com/go-chess-backend/internal/auth"
	"krishna.com/go-chess-backend/internal/game"
	"krishna.com/go-chess-backend/internal/store"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return r.Header.Get("Origin") == "http://localhost:5173"
	},
}

func ServeWS(h *Hub, w http.ResponseWriter, r *http.Request, st store.Store, manager *game.GameManager) {
	claims, err := auth.TokenFromRequest(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	playerID := claims.UserID

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("ws upgrade:", err)
		return
	}

	var initMsg Message
	if err := conn.ReadJSON(&initMsg); err != nil {
		_ = conn.Close()
		return
	}

	if initMsg.Type != "auth" {
		_ = conn.Close()
		return
	}

	var payload struct {
		GameID string `json:"gameId"`
	}
	json.Unmarshal(initMsg.Payload, &payload)

	_ = conn.SetReadDeadline(time.Now().Add(24 * time.Hour))
	conn.SetPongHandler(func(string) error {
		_ = conn.SetReadDeadline(time.Now().Add(24 * time.Hour))
		return nil
	})

	if payload.GameID == "" {
		serveLobby(h, conn, playerID)
		return
	}

	serveGame(h, conn, st, manager, playerID, payload.GameID)
}

func serveLobby(h *Hub, conn *websocket.Conn, playerID string) {
	h.RegisterLobby(playerID, conn)
	defer h.UnregisterLobby(playerID, conn)

	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}
}

func serveGame(h *Hub, conn *websocket.Conn, st store.Store, manager *game.GameManager, playerID, gameID string) {
	h.RegisterGame(playerID, gameID, conn)
	defer h.UnregisterGame(playerID, gameID, conn)

	conn.SetReadLimit(4096)

	for {
		var msg Message
		if err := conn.ReadJSON(&msg); err != nil {
			fmt.Println("ws read error:", err)
			break
		}

		if msg.Type == "move" {
			log.Println("move message", msg)
			var movePayload MovePayload
			json.Unmarshal(msg.Payload, &movePayload)
			log.Println("move:", movePayload.Move)

			st.ApplyMove(msg.GameID, movePayload.Move, playerID)

			opponentID, ok := opponentOf(manager, st, msg.GameID, playerID)
			if !ok {
				continue
			}
			h.SendToGame(opponentID, msg.GameID, msg)

			if outcome := st.Outcome(msg.GameID); outcome != "" {
				broadcastGameOver(h, msg.GameID, outcome, playerID, opponentID)
			}
		}

		if msg.Type == "resign" {
			if err := st.Resign(msg.GameID, playerID); err != nil {
				log.Println("resign:", err)
				continue
			}
			opponentID, ok := opponentOf(manager, st, msg.GameID, playerID)
			if !ok {
				continue
			}
			outcome := st.Outcome(msg.GameID)
			broadcastGameOver(h, msg.GameID, outcome, playerID, opponentID)
		}
	}
}

func opponentOf(manager *game.GameManager, st store.Store, gameID, playerID string) (string, bool) {
	var playerA, playerB string
	if g := manager.GetGame(gameID); g != nil {
		playerA, playerB = g.PlayerA, g.PlayerB
	} else {
		gm, err := st.GetGame(gameID)
		if err != nil {
			log.Println("could not get the game:", gameID, err)
			return "", false
		}
		playerA, playerB = gm.PlayerA, gm.PlayerB
	}
	if playerID == playerA {
		return playerB, true
	}
	return playerA, true
}

func broadcastGameOver(h *Hub, gameID, outcome, playerA, playerB string) {
	msg := Message{
		Type:    "game_over",
		GameID:  gameID,
		Payload: json.RawMessage(fmt.Sprintf(`{"outcome":"%s"}`, outcome)),
	}
	h.SendToGame(playerA, gameID, msg)
	h.SendToGame(playerB, gameID, msg)
}
