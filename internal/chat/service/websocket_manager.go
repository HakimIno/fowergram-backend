package service

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/fasthttp/websocket"
	"github.com/redis/go-redis/v9"
)

type Client struct {
	conn     *websocket.Conn
	lastPing time.Time
	isActive bool
	mu       sync.RWMutex
}

type WebSocketManager struct {
	clients    map[string]*Client
	mutex      sync.RWMutex
	redisCache *redis.Client
	done       chan struct{}
}

func NewWebSocketManager(redisClient *redis.Client) *WebSocketManager {
	manager := &WebSocketManager{
		clients:    make(map[string]*Client),
		redisCache: redisClient,
		done:       make(chan struct{}),
	}

	go manager.subscribeToRedis()
	go manager.cleanInactiveConnections()
	return manager
}

func (m *WebSocketManager) cleanInactiveConnections() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.mutex.Lock()
			for userID, client := range m.clients {
				client.mu.RLock()
				if !client.isActive || time.Since(client.lastPing) > 2*time.Minute {
					client.conn.Close()
					delete(m.clients, userID)
				}
				client.mu.RUnlock()
			}
			m.mutex.Unlock()
		case <-m.done:
			return
		}
	}
}

func (m *WebSocketManager) subscribeToRedis() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pubsub := m.redisCache.Subscribe(ctx, "chat:*")
	defer pubsub.Close()

	ch := pubsub.Channel()
	for {
		select {
		case msg := <-ch:
			m.broadcastMessage([]byte(msg.Payload))
		case <-m.done:
			return
		}
	}
}

func (m *WebSocketManager) broadcastMessage(message []byte) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	for _, client := range m.clients {
		client.mu.RLock()
		if client.isActive {
			go func(c *Client) {
				if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
					c.mu.Lock()
					c.isActive = false
					c.mu.Unlock()
				}
			}(client)
		}
		client.mu.RUnlock()
	}
}

func (m *WebSocketManager) AddClient(userID string, conn *websocket.Conn) {
	client := &Client{
		conn:     conn,
		lastPing: time.Now(),
		isActive: true,
	}

	m.mutex.Lock()
	m.clients[userID] = client
	m.mutex.Unlock()

	go m.handleConnection(userID, client)
}

func (m *WebSocketManager) RemoveClient(userID string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if client, exists := m.clients[userID]; exists {
		client.mu.Lock()
		client.isActive = false
		client.conn.Close()
		client.mu.Unlock()
		delete(m.clients, userID)
	}
}

func (m *WebSocketManager) SendToUser(userID string, message interface{}) error {
	m.mutex.RLock()
	client, exists := m.clients[userID]
	m.mutex.RUnlock()

	if !exists || !client.isActive {
		return nil
	}

	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	client.mu.Lock()
	defer client.mu.Unlock()
	return client.conn.WriteMessage(websocket.TextMessage, data)
}

func (m *WebSocketManager) handleConnection(userID string, client *Client) {
	pingTicker := time.NewTicker(30 * time.Second)
	defer pingTicker.Stop()

	go func() {
		for range pingTicker.C {
			client.mu.RLock()
			if !client.isActive {
				client.mu.RUnlock()
				return
			}
			client.mu.RUnlock()

			client.mu.Lock()
			err := client.conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(10*time.Second))
			client.mu.Unlock()
			if err != nil {
				client.mu.Lock()
				client.isActive = false
				client.mu.Unlock()
				return
			}
		}
	}()

	client.conn.SetPingHandler(func(string) error {
		client.mu.Lock()
		client.lastPing = time.Now()
		client.mu.Unlock()
		return client.conn.WriteControl(websocket.PongMessage, nil, time.Now().Add(10*time.Second))
	})

	// Set initial read deadline
	client.conn.SetReadDeadline(time.Now().Add(2 * time.Minute))
}

func (m *WebSocketManager) Shutdown() {
	close(m.done)

	m.mutex.Lock()
	defer m.mutex.Unlock()

	for _, client := range m.clients {
		client.mu.Lock()
		client.isActive = false
		client.conn.Close()
		client.mu.Unlock()
	}
	m.clients = make(map[string]*Client)
}
