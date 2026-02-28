package websocket

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

// Event types
const (
	EventAgentStatus  = "agent.status"
	EventTaskStatus   = "task.status"
	EventPhaseUpdated = "phase.updated"
	EventStoryUpdated = "story.updated"
	EventNewEvent     = "event.new"
	EventExecutionLog = "execution.log"
)

type Message struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

type Client struct {
	hub  *Hub
	conn *websocket.Conn
	send chan []byte
}

type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			log.Printf("WebSocket client connected. Total: %d", len(h.clients))

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()
			log.Printf("WebSocket client disconnected. Total: %d", len(h.clients))

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// Broadcast sends a message to all connected clients
func (h *Hub) Broadcast(msg *Message) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
		return
	}
	h.broadcast <- data
}

// BroadcastAgentStatus sends agent status update
func (h *Hub) BroadcastAgentStatus(agentID, status string, taskID *string) {
	h.Broadcast(&Message{
		Type: EventAgentStatus,
		Payload: map[string]interface{}{
			"agent_id":        agentID,
			"status":          status,
			"current_task_id": taskID,
		},
	})
}

// BroadcastTaskStatus sends task status update
func (h *Hub) BroadcastTaskStatus(taskID, status string, progress float64) {
	h.Broadcast(&Message{
		Type: EventTaskStatus,
		Payload: map[string]interface{}{
			"task_id":  taskID,
			"status":   status,
			"progress": progress,
		},
	})
}

// BroadcastEvent sends a new event notification
func (h *Hub) BroadcastEvent(event interface{}) {
	h.Broadcast(&Message{
		Type:    EventNewEvent,
		Payload: event,
	})
}

// Client methods
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			break
		}
		// We don't process incoming messages for now
		// Could add ping/pong handling here
	}
}

func (c *Client) writePump() {
	defer c.conn.Close()

	for message := range c.send {
		if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
			return
		}
	}
}

// RegisterClient creates a new client and starts its read/write pumps
func (h *Hub) RegisterClient(conn *websocket.Conn) {
	client := &Client{
		hub:  h,
		conn: conn,
		send: make(chan []byte, 256),
	}

	h.register <- client

	go client.writePump()
	go client.readPump()
}
