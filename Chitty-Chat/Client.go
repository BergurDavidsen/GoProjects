package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"sync/atomic"

	"google.golang.org/grpc"
)

type ChatClient struct {
	stream    Services_ChatClient
	clientID  string
	name      string
	timestamp int64
}

func NewChatClient(stream Services_ChatClient, name string) *ChatClient {
	return &ChatClient{
		stream:    stream,
		name:      name,
		timestamp: 0,
	}
}

func (c *ChatClient) updateTimestamp(serverTime int64) int64 {
	for {
		current := atomic.LoadInt64(&c.timestamp)
		new := max(current, serverTime) + 1
		if atomic.CompareAndSwapInt64(&c.timestamp, current, new) {
			return new
		}
	}
}

func (c *ChatClient) joinChat() error {
	timestamp := atomic.AddInt64(&c.timestamp, 1)
	return c.stream.Send(&FromClient{
		Name:      c.name,
		Type:      MessageType_JOIN,
		Timestamp: timestamp,
	})
}

func (c *ChatClient) sendMessage(message string) error {
	if len(message) > 128 {
		return fmt.Errorf("message too long (max 128 characters)")
	}

	timestamp := atomic.AddInt64(&c.timestamp, 1)
	return c.stream.Send(&FromClient{
		Name:      c.name,
		Body:      message,
		Type:      MessageType_CHAT_MESSAGE,
		Timestamp: timestamp,
	})
}

func (c *ChatClient) receiveMessages() {
	for {
		msg, err := c.stream.Recv()
		if err != nil {
			log.Printf("Error receiving message: %v", err)
			return
		}

		// Update local timestamp
		c.updateTimestamp(msg.Timestamp)

		// Log message with timestamp
		switch msg.Type {
		case MessageType_JOIN:
			log.Printf("[T=%d] %s", msg.Timestamp, msg.Body)
		case MessageType_LEAVE:
			log.Printf("[T=%d] %s", msg.Timestamp, msg.Body)
		case MessageType_CHAT_MESSAGE:
			log.Printf("[T=%d] %s: %s", msg.Timestamp, msg.Name, msg.Body)
		}
	}
}

func main() {
	// Get server address from user
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter server address (host:port): ")
	serverAddr, _ := reader.ReadString('\n')
	serverAddr = strings.TrimSpace(serverAddr)

	// Connect to server
	conn, err := grpc.Dial(serverAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	// Create chat service client
	client := NewServicesClient(conn)
	stream, err := client.Chat(context.Background())
	if err != nil {
		log.Fatalf("Failed to create chat stream: %v", err)
	}

	// Get user's name
	fmt.Print("Enter your name: ")
	name, _ := reader.ReadString('\n')
	name = strings.TrimSpace(name)

	// Create chat client
	chatClient := NewChatClient(stream, name)

	// Join chat
	if err := chatClient.joinChat(); err != nil {
		log.Fatalf("Failed to join chat: %v", err)
	}

	// Start receiving messages in background
	go chatClient.receiveMessages()

	// Handle user input
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("Error reading input: %v", err)
			break
		}

		message = strings.TrimSpace(message)
		if message == "/quit" {
			break
		}

		if len(message) > 128 {
			log.Printf("Error: message too long (max 128 characters)")
			continue
		}

		if err := chatClient.sendMessage(message); err != nil {
			log.Printf("Error sending message: %v", err)
		}
	}
