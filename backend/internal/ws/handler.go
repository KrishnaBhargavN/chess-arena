package ws

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"krishna.com/go-chess-backend/internal/game"
	"krishna.com/go-chess-backend/internal/store"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func ServeWS(h *Hub, w http.ResponseWriter, r *http.Request, st *store.MemoryStore, manager *game.GameManager) {
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
		PlayerID string `json:"playerId"`
		GameID   string `json:"gameId"`
	}
	json.Unmarshal(initMsg.Payload, &payload)

	if payload.PlayerID == "" {
		_ = conn.Close()
		return
	}

	_ = conn.SetReadDeadline(time.Now().Add(24 * time.Hour))
	conn.SetPongHandler(func(string) error {
		_ = conn.SetReadDeadline(time.Now().Add(24 * time.Hour))
		return nil
	})

	if payload.GameID == "" {
		serveLobby(h, conn, payload.PlayerID)
		return
	}

	serveGame(h, conn, st, manager, payload.PlayerID, payload.GameID)
}

// serveLobby handles a connection that exists only to receive matchmaking notifications.
func serveLobby(h *Hub, conn *websocket.Conn, playerID string) {
	h.RegisterLobby(playerID, conn)
	defer h.UnregisterLobby(playerID, conn)

	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}
}

// serveGame handles a connection scoped to a specific game, forwarding move messages.
func serveGame(h *Hub, conn *websocket.Conn, st *store.MemoryStore, manager *game.GameManager, playerID, gameID string) {
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

			st.ApplyMove(msg.GameID, movePayload.Move, movePayload.PlayerId)

			g := manager.GetGame(msg.GameID)
			if g == nil {
				log.Println("could not get the game:", msg.GameID)
				continue
			}

			var opponentID string
			if movePayload.PlayerId == g.PlayerA {
				opponentID = g.PlayerB
			} else {
				opponentID = g.PlayerA
			}
			h.SendToGame(opponentID, msg.GameID, msg)
		}
	}
}
