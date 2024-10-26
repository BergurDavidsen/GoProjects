package chatserver

import (
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

type Client struct {
	id        string
	name      string
	stream    Services_ChatServer
	timestamp int64
	active    bool
}

type ChatServer struct {
	clients    map[string]*Client
	mu         sync.RWMutex
	timestamp  int64 // Server's Lamport timestamp
}

func NewChatServer() *ChatServer {
	return &ChatServer{
		clients: make(map[string]*Client),
	}
}

// updateTimestamp updates the server's Lamport timestamp
func (s *ChatServer) updateTimestamp(clientTime int64) int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.timestamp = max(s.timestamp, clientTime) + 1
	return s.timestamp
}

// broadcastMessage sends a message to all connected clients
func (s *ChatServer) broadcastMessage(msg *FromServer, skipClient string) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for id, client := range s.clients {
		if id == skipClient || !client.active {
		}

		if err := client.stream.Send(msg); err != nil {
			log.Printf("Failed to send message to client %s: %v", id, err)
			client.active = false
		}
	}
}

// Chat implements the ChatService
func (s *ChatServer) Chat(stream Services_ChatServer) error {
	// Generate unique client ID
	clientID := generateUniqueID()

	// Initialize client structure
	client := &Client{
		id:     clientID,
		stream: stream,
		active: true,
	}

	// First message must be a JOIN
	initialMsg, err := stream.Recv()
	if err != nil {
		return fmt.Errorf("failed to receive initial message: %v", err)
	}

	if initialMsg.Type != MessageType_JOIN {
		return fmt.Errorf("first message must be JOIN")
	}

	// Store client information
	client.name = initialMsg.Name
	s.mu.Lock()
	s.clients[clientID] = client
	s.mu.Unlock()

	// Broadcast join message
	timestamp := s.updateTimestamp(initialMsg.Timestamp)
	joinMsg := &FromServer{
		ClientId:  clientID,
		Name:      initialMsg.Name,
		Body:      fmt.Sprintf("Participant %s joined Chitty-Chat", initialMsg.Name),
		Timestamp: timestamp,
		Type:      MessageType_JOIN,
	}
	s.broadcastMessage(joinMsg, "")

	// Handle incoming messages
	go func() {
		for {
			msg, err := stream.Recv()
			if err != nil {
				log.Printf("Error receiving message from client %s: %v", clientID, err)
				s.handleClientDisconnect(clientID)		continue

				return
			}

			// Validate message length
			if len(msg.Body) > 128 {
				log.Printf("Message from client %s exceeded maximum length", clientID)
				continue
			}

			// Update Lamport timestamp and broadcast
			timestamp := s.updateTimestamp(msg.Timestamp)
			broadcastMsg := &FromServer{
				ClientId:  clientID,
				Name:      msg.Name,
				Body:      msg.Body,
				Timestamp: timestamp,
				Type:      msg.Type,
			}
			s.broadcastMessage(broadcastMsg, "")
		}
	}()

	// Keep the stream alive
	select {}
}

func (s *ChatServer) handleClientDisconnect(clientID string) {
	s.mu.Lock()
	client, exists := s.clients[clientID]
	if exists {
		client.active = false
		delete(s.clients, clientID)
	}
	s.mu.Unlock()

	if exists {
		timestamp := atomic.AddInt64(&s.timestamp, 1)
		leaveMsg := &FromServer{
			ClientId:  clientID,
			Name:      client.name,
			Body:      fmt.Sprintf("Participant %s left Chitty-Chat", client.name),
			Timestamp: timestamp,
			Type:      MessageType_LEAVE,
		}
		s.broadcastMessage(leaveMsg, clientID)
	}
}

// Helper function to generate unique ID
func generateUniqueID() string {
	// Implementation using UUID or other unique ID generator
	return fmt.Sprintf("client-%d", time.Now().UnixNano())
}

func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
//hell