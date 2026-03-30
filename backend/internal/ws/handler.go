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
		log.Println("ws upgrade: ", err)
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
	}

	json.Unmarshal(initMsg.Payload, &payload)

	if payload.PlayerID == "" {
		conn.Close()
		return
	}

	h.Register(payload.PlayerID, conn)
	defer h.Unregister(payload.PlayerID)

	conn.SetReadLimit(512)
	_ = conn.SetReadDeadline(time.Now().Add(24 * time.Hour))
	conn.SetPongHandler(func(string) error {
		_ = conn.SetReadDeadline(time.Now().Add(24 * time.Hour))
		return nil
	})

	for {
		var msg Message
		if err := conn.ReadJSON(&msg); err != nil {
			fmt.Println("here ", err)
			break
		}

		if msg.Type == "move" {
			log.Println("move message", msg)
			var movePayload MovePayload
			json.Unmarshal(msg.Payload, &movePayload)
			log.Println("move: ", movePayload.Move)
			st.ApplyMove(msg.GameID, movePayload.Move, movePayload.PlayerId)
			game := manager.GetGame(msg.GameID)
			if game == nil {
				log.Println("could not get the game")
				return
			}

			if movePayload.PlayerId == game.PlayerA {
				h.SendTo(game.PlayerB, msg)
			} else {
				h.SendTo(game.PlayerA, msg)
			}
		}
	}
}
