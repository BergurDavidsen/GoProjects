package main

import (
    "log"
    "math/rand"
    "net"
    "os"
    "sync"
    "time"
    "google.golang.org/grpc"
    "grpcChatServer/chatserver"
)

// MessageUnit represents a single message in the chat system
type messageUnit struct {
    ClientName        string
    MessageBody       string
    MessageUniqueCode int
    ClientUniqueCode  int
    IsSystemMessage   bool
    Timestamp        string
}

// MessageHandle manages the message queue with thread-safe operations
type messageHandle struct {
    MQue []messageUnit
    mu   sync.Mutex
}

// Global variables and constants
var messageHandleObject = messageHandle{}
const MaxMessageLength = 128

// ChatServer implements the gRPC service
type ChatServer struct {
    chatserver.UnimplementedServicesServer
}

// ChatService implements the bi-directional streaming RPC for chat
func (cs *ChatServer) ChatService(csi chatserver.Services_ChatServiceServer) error {
    clientUniqueCode := rand.Intn(1e6)
    errch := make(chan error)

    go receiveFromStream(csi, clientUniqueCode, errch)
    go sendToStream(csi, clientUniqueCode, errch)

    return <-errch
}

// getCurrentTimestamp returns the current time in a formatted string
func getCurrentTimestamp() string {
    return time.Now().Format("15:04:05")
}

// receiveFromStream handles incoming messages from clients
func receiveFromStream(csi chatserver.Services_ChatServiceServer, clientUniqueCode int, errch chan error) {
    for {
        mssg, err := csi.Recv()
        if err != nil {
            log.Printf("Error in receiving message from client: %v", err)
            errch <- err
            return
        }

        timestamp := getCurrentTimestamp()

        // Check message length
        if len(mssg.Body) > MaxMessageLength {
            log.Printf("[%s] Client %s exceeded message length limit", timestamp, mssg.Name)
            errorMessage := messageUnit{
                ClientName:        "System",
                MessageBody:       "Exceeded character limit of 128, please write a smaller message",
                MessageUniqueCode: rand.Intn(1e8),
                ClientUniqueCode:  clientUniqueCode,
                IsSystemMessage:   true,
                Timestamp:        timestamp,
            }
            messageHandleObject.mu.Lock()
            messageHandleObject.MQue = append(messageHandleObject.MQue, errorMessage)
            messageHandleObject.mu.Unlock()
            continue
        }

        messageHandleObject.mu.Lock()
        messageHandleObject.MQue = append(messageHandleObject.MQue, messageUnit{
            ClientName:        mssg.Name,
            MessageBody:       mssg.Body,
            MessageUniqueCode: rand.Intn(1e8),
            ClientUniqueCode:  clientUniqueCode,
            IsSystemMessage:   false,
            Timestamp:        timestamp,
        })
        log.Printf("[%s] Received message from %s: %s", timestamp, mssg.Name, mssg.Body)
        messageHandleObject.mu.Unlock()
    }
}

// sendToStream handles outgoing messages to clients
func sendToStream(csi chatserver.Services_ChatServiceServer, clientUniqueCode int, errch chan error) {
    for {
        time.Sleep(500 * time.Millisecond)

        messageHandleObject.mu.Lock()
        if len(messageHandleObject.MQue) == 0 {
            messageHandleObject.mu.Unlock()
            continue
        }

        currentMessage := messageHandleObject.MQue[0]
        messageHandleObject.mu.Unlock()

        var err error
        if currentMessage.IsSystemMessage && currentMessage.ClientUniqueCode == clientUniqueCode {
            // Send system message only to the intended client
            err = csi.Send(&chatserver.FromServer{
                Name:           currentMessage.ClientName,
                Body:           currentMessage.MessageBody,
                IsSystemMessage: true,
                Timestamp:      currentMessage.Timestamp,
            })
        } else if !currentMessage.IsSystemMessage && currentMessage.ClientUniqueCode != clientUniqueCode {
            // Send regular message to all other clients
            err = csi.Send(&chatserver.FromServer{
                Name:           currentMessage.ClientName,
                Body:           currentMessage.MessageBody,
                IsSystemMessage: false,
                Timestamp:      currentMessage.Timestamp,
            })
        }

        if err != nil {
            errch <- err
            continue
        }

        // Remove the sent message from the queue
        messageHandleObject.mu.Lock()
        if len(messageHandleObject.MQue) > 0 {
            messageHandleObject.MQue = messageHandleObject.MQue[1:]
        }
        messageHandleObject.mu.Unlock()
    }
}

func main() {
    // Initialize random seed
    rand.Seed(time.Now().UnixNano())

    // Get port from environment variable or use default
    Port := os.Getenv("PORT")
    if Port == "" {
        Port = "5000"
    }

    // Initialize TCP listener
    listen, err := net.Listen("tcp", ":"+Port)
    if err != nil {
        log.Fatalf("Could not listen @ %v :: %v", Port, err)
    }
    log.Println("Server listening @ :" + Port)

    // Create and start gRPC server
    grpcserver := grpc.NewServer()
    chatserver.RegisterServicesServer(grpcserver, &ChatServer{})

    // Start serving
    if err := grpcserver.Serve(listen); err != nil {
        log.Fatalf("Failed to start gRPC Server :: %v", err)
    }
}