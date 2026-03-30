package ws

import (
	"encoding/json"
	"errors"
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
		mu:      sync.RWMutex{},
		clients: make(map[string]*websocket.Conn),
		pending: make(map[string]Message),
	}
}

func (h *Hub) Register(playerID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if old, ok := h.clients[playerID]; ok {
		_ = old.Close()
	}

	h.clients[playerID] = conn

	if msg, ok := h.pending[playerID]; ok {
		_ = conn.WriteJSON(msg)
		delete(h.pending, playerID)
	}
}

func (h *Hub) Unregister(playerID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if c, ok := h.clients[playerID]; ok {
		_ = c.Close()
		delete(h.clients, playerID)
	}
}

func (h *Hub) SendTo(playerID string, msg Message) error {
	h.mu.Lock()
	conn, ok := h.clients[playerID]
	h.mu.Unlock()

	if ok {
		if err := conn.WriteJSON(msg); err != nil {
			h.mu.Lock()
			delete(h.clients, playerID)
			h.pending[playerID] = msg
			h.mu.Unlock()
			return err
		}
		return nil
	}

	h.mu.Lock()
	h.pending[playerID] = msg
	h.mu.Unlock()
	return errors.New("no-conn, stored pending")
}
