package service

import (
	"encoding/json"
	"sync"

	"fowergram/internal/domain/chat"

	"github.com/gofiber/websocket/v2"
)

type WebSocketManager struct {
	connections sync.Map // map[string]map[string]*websocket.Conn // userID -> connectionID -> connection
	mu          sync.RWMutex
}

func NewWebSocketManager() *WebSocketManager {
	return &WebSocketManager{}
}

func (m *WebSocketManager) AddConnection(userID string, conn *websocket.Conn) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if conns, ok := m.connections.Load(userID); ok {
		if userConns, ok := conns.(map[string]*websocket.Conn); ok {
			userConns[conn.LocalAddr().String()] = conn
		} else {
			userConns = make(map[string]*websocket.Conn)
			userConns[conn.LocalAddr().String()] = conn
			m.connections.Store(userID, userConns)
		}
	} else {
		userConns := make(map[string]*websocket.Conn)
		userConns[conn.LocalAddr().String()] = conn
		m.connections.Store(userID, userConns)
	}
}

func (m *WebSocketManager) RemoveConnection(userID string, conn *websocket.Conn) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if conns, ok := m.connections.Load(userID); ok {
		if userConns, ok := conns.(map[string]*websocket.Conn); ok {
			delete(userConns, conn.LocalAddr().String())
			if len(userConns) == 0 {
				m.connections.Delete(userID)
			}
		}
	}
}

func (m *WebSocketManager) BroadcastToChat(chatID string, message *domain.Message) {
	m.connections.Range(func(key, value interface{}) bool {
		userID := key.(string)
		if userConns, ok := value.(map[string]*websocket.Conn); ok {
			for _, conn := range userConns {
				go func(c *websocket.Conn) {
					if err := m.sendMessage(c, message); err != nil {
						// Handle error (maybe remove connection if it's dead)
						m.RemoveConnection(userID, c)
					}
				}(conn)
			}
		}
		return true
	})
}

func (m *WebSocketManager) SendToUser(userID string, message *domain.Message) error {
	if conns, ok := m.connections.Load(userID); ok {
		if userConns, ok := conns.(map[string]*websocket.Conn); ok {
			for _, conn := range userConns {
				if err := m.sendMessage(conn, message); err != nil {
					// If one connection fails, try others
					continue
				}
				return nil
			}
		}
	}
	return nil // User might be offline, which is not an error
}

func (m *WebSocketManager) sendMessage(conn *websocket.Conn, message *domain.Message) error {
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}
	return conn.WriteMessage(websocket.TextMessage, data)
}

func (m *WebSocketManager) Shutdown() {
	m.connections.Range(func(key, value interface{}) bool {
		if userConns, ok := value.(map[string]*websocket.Conn); ok {
			for _, conn := range userConns {
				conn.Close()
			}
		}
		return true
	})
}
