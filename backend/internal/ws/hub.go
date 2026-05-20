package ws

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/gorilla/websocket"
)

type Message struct {
	Type    string          `json:"type"`
	GameID  string          `json:"gameId,omitempty"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

type MovePayload struct {
	GameID   string `json:"gameId"`
	Move     string `json:"move"`
	From     string `json:"from"`
	To       string `json:"to"`
	PlayerId string `json:"playerId"`
}

type Hub struct {
	mu      sync.RWMutex
	clients map[string]*websocket.Conn
	pending map[string]Message
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[string]*websocket.Conn),
		pending: make(map[string]Message),
	}
}

func gameKey(playerID, gameID string) string {
	return fmt.Sprintf("%s:%s", playerID, gameID)
}

func (h *Hub) RegisterLobby(playerID string, conn *websocket.Conn) {
	h.register(playerID, conn)
}

func (h *Hub) RegisterGame(playerID, gameID string, conn *websocket.Conn) {
	h.register(gameKey(playerID, gameID), conn)
}

func (h *Hub) UnregisterLobby(playerID string, conn *websocket.Conn) {
	h.unregister(playerID, conn)
}

func (h *Hub) UnregisterGame(playerID, gameID string, conn *websocket.Conn) {
	h.unregister(gameKey(playerID, gameID), conn)
}

func (h *Hub) SendToLobby(playerID string, msg Message) error {
	return h.sendTo(playerID, msg)
}

func (h *Hub) SendToGame(playerID, gameID string, msg Message) error {
	return h.sendTo(gameKey(playerID, gameID), msg)
}

func (h *Hub) register(key string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if old, ok := h.clients[key]; ok {
		_ = old.Close()
	}
	h.clients[key] = conn

	if msg, ok := h.pending[key]; ok {
		_ = conn.WriteJSON(msg)
		delete(h.pending, key)
	}
}

func (h *Hub) unregister(key string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if c, ok := h.clients[key]; ok && c == conn {
		_ = c.Close()
		delete(h.clients, key)
	}
}

func (h *Hub) sendTo(key string, msg Message) error {
	h.mu.Lock()
	conn, ok := h.clients[key]
	h.mu.Unlock()

	if ok {
		if err := conn.WriteJSON(msg); err != nil {
			h.mu.Lock()
			delete(h.clients, key)
			h.pending[key] = msg
			h.mu.Unlock()
			return err
		}
		return nil
	}

	h.mu.Lock()
	h.pending[key] = msg
	h.mu.Unlock()
	return errors.New("no-conn, stored pending")
}
