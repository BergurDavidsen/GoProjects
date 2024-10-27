package main

import (
	"grpcChatServer/chatserver"
	"log"
	"math/rand"
	"net"
	"os"
	"sync"
	"time"

	"google.golang.org/grpc"
)

// MessageUnit represents a single message in the chat system

// TODO: Add a Lamport timestamp to the messageUnit struct
type messageUnit struct {
	ClientName        string
	MessageBody       string
	MessageUniqueCode int
	ClientUniqueCode  int
	IsSystemMessage   bool
	Timestamp         string
	LamportTimestamp  uint32
}

// MessageHandle manages the message queue with thread-safe operations
type messageHandle struct {
	MQue []messageUnit
	mu   sync.Mutex
}

// Global variables and constants
var messageHandleObject = messageHandle{}
var LamportTimestamp uint32 = 0

const MaxMessageLength = 128

// ChatServer implements the gRPC service
type ChatServer struct {
	chatserver.UnimplementedServicesServer
	clients map[chatserver.Services_ChatServiceServer]bool // Track connected clients
	mu      sync.Mutex
}

// ChatService implements the bi-directional streaming RPC for chat
func (cs *ChatServer) ChatService(csi chatserver.Services_ChatServiceServer) error {
	errch := make(chan error)

	// Add client to the map
	cs.mu.Lock()
	LamportTimestamp++
	log.Printf("Participant %s joined Chitty-Chat at Lamport time %d", csi.Context().Value("name"), LamportTimestamp)
	cs.clients[csi] = true
	cs.mu.Unlock()

	go receiveFromStream(csi, cs, errch) // Pass cs to receiveFromStream
	go cs.sendToStream()



	// Wait for error
	return <-errch
}

// getCurrentTimestamp returns the current time in a formatted string

func getCurrentTimestamp() string {
	return time.Now().Format("15:04:05")
}

// receiveFromStream handles incoming messages from clients
func receiveFromStream(csi chatserver.Services_ChatServiceServer, chatServer *ChatServer, errch chan error) {
	defer func() {
		// Clean up the client when it disconnects
		chatServer.mu.Lock()
		delete(chatServer.clients, csi)
		chatServer.mu.Unlock()
	}()

	for {
		mssg, err := csi.Recv()
		if err != nil {
			log.Printf("Error in receiving message from client: %v", err)
			errch <- err
			return
		}

		timestamp := getCurrentTimestamp()
		messageHandleObject.mu.Lock()
		LamportTimestamp = (max(LamportTimestamp, mssg.LamportTimestamp) + 1)
		messageHandleObject.mu.Unlock()

		// Check message length
		if len(mssg.Body) > MaxMessageLength {
			log.Printf("[%s] Client %s exceeded message length limit", timestamp, mssg.Name)
			errorMessage := messageUnit{
				ClientName:        "System",
				MessageBody:       "Exceeded character limit of 128, please write a smaller message",
				MessageUniqueCode: rand.Intn(1e8),
				ClientUniqueCode:  clientUniqueCode,
				IsSystemMessage:   true,
				Timestamp:         timestamp,
				LamportTimestamp:  LamportTimestamp,
			}
			messageHandleObject.mu.Lock()
			messageHandleObject.MQue = append(messageHandleObject.MQue, errorMessage)
			messageHandleObject.mu.Unlock()
			continue
		}

		messageHandleObject.mu.Lock()
		//TODO: Add Lamport timestamp evaluation here

		messageHandleObject.MQue = append(messageHandleObject.MQue, messageUnit{
			ClientName:        mssg.Name,
			MessageBody:       mssg.Body,
			MessageUniqueCode: rand.Intn(1e8),
			ClientUniqueCode:  clientUniqueCode,
			IsSystemMessage:   false,
			Timestamp:         timestamp,
			LamportTimestamp:  LamportTimestamp,
		})
		log.Printf("{%d}[%s] Received message from %s: %s", LamportTimestamp, timestamp, mssg.Name, mssg.Body)
		messageHandleObject.mu.Unlock()
	}
}

// sendToStream handles outgoing messages to all clients
func (cs *ChatServer) sendToStream() {
	for {
		time.Sleep(500 * time.Millisecond) // Control message sending rate

		messageHandleObject.mu.Lock()
		if len(messageHandleObject.MQue) == 0 {
			messageHandleObject.mu.Unlock()
			continue
		}

		// Get the next message to send
		currentMessage := messageHandleObject.MQue[0]
		messageHandleObject.mu.Unlock()

		// Prepare the message to send
		serverMessage := &chatserver.FromServer{
			Name:             currentMessage.ClientName,
			Body:             currentMessage.MessageBody,
			IsSystemMessage:  currentMessage.IsSystemMessage,
			Timestamp:        currentMessage.Timestamp,
			LamportTimestamp: LamportTimestamp,
		}

		// Broadcast message to all clients
		cs.mu.Lock()
		for client := range cs.clients {
			// Send the message to all clients
			LamportTimestamp++
			serverMessage.LamportTimestamp = LamportTimestamp
			err := client.Send(serverMessage)
			if err != nil {
				log.Printf("Failed to send message to client: %v", err)
				delete(cs.clients, client) // Remove the client if there's an error
			}
		}
		cs.mu.Unlock()

		// Remove the sent message from the queue
		messageHandleObject.mu.Lock()
		if len(messageHandleObject.MQue) > 0 {
			messageHandleObject.MQue = messageHandleObject.MQue[1:] // Remove the first message
		}
		messageHandleObject.mu.Unlock()
	}
}

func main() {

	// Get port from environment variable or use default
	Port := os.Getenv("PORT")
	if Port == "" {
		Port = "5001"
	}

	// Initialize TCP listener
	listen, err := net.Listen("tcp", ":"+Port)
	if err != nil {
		log.Fatalf("Could not listen @ %v :: %v", Port, err)
	}
	log.Println("Server listening @ :" + Port)

	// Create and start gRPC server
	grpcserver := grpc.NewServer()

	// Initialize ChatServer with an empty clients map
	chatServer := &ChatServer{
		clients: make(map[chatserver.Services_ChatServiceServer]bool),
	}

	chatserver.RegisterServicesServer(grpcserver, chatServer)

	// Start serving
	if err := grpcserver.Serve(listen); err != nil {
		log.Fatalf("Failed to start gRPC Server :: %v", err)
	}
}
