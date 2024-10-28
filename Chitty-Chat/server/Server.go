package main

import (
	"fmt"
	"grpcChatServer/chatserver"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc"
)

// messageUnit represents a single message with client and system details.
type messageUnit struct {
	ClientName        string
	MessageBody       string
	MessageUniqueCode int
	ClientUniqueCode  string
	IsSystemMessage   bool
	Timestamp         string
	LamportTimestamp  uint32
}

// messageHandle manages the message queue with mutex locks for thread safety.
type messageHandle struct {
	MQue []messageUnit
	mu   sync.Mutex
}

var messageHandleObject = messageHandle{}
var LamportTimestamp uint32 = 1
const MaxMessageLength = 128

// ChatServer is the core server struct that handles client connections and metadata.
type ChatServer struct {
	chatserver.UnimplementedServicesServer
	clients        map[chatserver.Services_ChatServiceServer]bool
	mu             sync.Mutex
	clientMetaData map[chatserver.Services_ChatServiceServer]metadata.MD
}

// ChatService manages the client's connection, metadata, and spawns goroutines for receiving and sending messages.
func (cs *ChatServer) ChatService(csi chatserver.Services_ChatServiceServer) error {
	errch := make(chan error)
	cs.mu.Lock()
	md, ok := metadata.FromIncomingContext(csi.Context())

	if !ok {
		log.Println("No metadata found in context")
	} else {
		cs.clientMetaData[csi] = md
		clientName := md["clientname"]
		clientId := md["clientid"]
		clientLamportTimestamp := md["lamport"]
		clientLamportTimestampInt, err := strconv.ParseUint(clientLamportTimestamp[0], 10, 32)

		if err != nil {
			log.Println("Error converting lamport timestamp to int")
		}

		LamportTimestamp = max(LamportTimestamp, uint32(clientLamportTimestampInt)) + 1

		if len(clientName) > 0 {
			message := messageUnit{
				MessageBody:       fmt.Sprintf("Participant %s joined Chitty-Chat at Lamport time %d\n", clientName[0], LamportTimestamp),
				MessageUniqueCode: rand.Intn(1e8),
				ClientUniqueCode:  clientId[0],
				IsSystemMessage:   true,
				Timestamp:         getCurrentTimestamp(),
				LamportTimestamp:  LamportTimestamp,
			}
			log.Printf(message.MessageBody)
			messageHandleObject.MQue = append(messageHandleObject.MQue, message)
		} else {
			log.Println("clientId not found in metadata")
		}
	}

	cs.clients[csi] = true
	cs.mu.Unlock()

	go receiveFromStream(csi, cs, errch)
	go cs.sendToStream()

	return <-errch
}

// getCurrentTimestamp retrieves the current time in a specific format.
func getCurrentTimestamp() string {
	return time.Now().Format("15:04:05")
}

// receiveFromStream continuously receives messages from clients and handles client disconnection.
func receiveFromStream(csi chatserver.Services_ChatServiceServer, chatServer *ChatServer, errch chan error) {
	defer func() {
		chatServer.mu.Lock()
		delete(chatServer.clients, csi)
		chatServer.mu.Unlock()
	}()

	for {
		mssg, err := csi.Recv()
		if err != nil {
			log.Printf("Error in receiving message from client: %v", err)
			LamportTimestamp++
			disconnectMessage := messageUnit{
				ClientName:       chatServer.clientMetaData[csi]["clientname"][0],
				MessageBody:      fmt.Sprintf("Participant %s left Chitty-Chat at Lamport time %d\n", chatServer.clientMetaData[csi]["clientname"][0], LamportTimestamp),
				Timestamp:        getCurrentTimestamp(),
				LamportTimestamp: LamportTimestamp,
				IsSystemMessage:  true,
			}
			log.Printf(disconnectMessage.MessageBody)
			messageHandleObject.mu.Lock()
			messageHandleObject.MQue = append(messageHandleObject.MQue, disconnectMessage)
			messageHandleObject.mu.Unlock()
			errch <- err
			return
		}

		timestamp := getCurrentTimestamp()
		messageHandleObject.mu.Lock()
		LamportTimestamp = max(LamportTimestamp, mssg.LamportTimestamp) + 1
		messageHandleObject.mu.Unlock()

		if len(mssg.Body) > MaxMessageLength {
			log.Printf("[%s] Client %s exceeded message length limit", timestamp, mssg.Name)
			errorMessage := messageUnit{
				ClientName:        "System",
				MessageBody:       fmt.Sprintf("Exceeded character limit of 128, please write a smaller message at Lamport time %d", LamportTimestamp),
				MessageUniqueCode: rand.Intn(1e8),
				ClientUniqueCode:  chatServer.clientMetaData[csi]["clientid"][0],
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
		messageHandleObject.MQue = append(messageHandleObject.MQue, messageUnit{
			ClientName:        mssg.Name,
			MessageBody:       mssg.Body,
			MessageUniqueCode: rand.Intn(1e8),
			ClientUniqueCode:  chatServer.clientMetaData[csi]["clientid"][0],
			IsSystemMessage:   false,
			Timestamp:         timestamp,
			LamportTimestamp:  LamportTimestamp,
		})
		log.Printf("[%s] Received message from %s at Lamport time %d: %s", timestamp, mssg.Name, LamportTimestamp, mssg.Body)
		messageHandleObject.mu.Unlock()
	}
}

// sendToStream loops through the message queue and broadcasts messages to all clients.
func (cs *ChatServer) sendToStream() {
	for {
		time.Sleep(500 * time.Millisecond)

		messageHandleObject.mu.Lock()
		if len(messageHandleObject.MQue) == 0 {
			messageHandleObject.mu.Unlock()
			continue
		}

		currentMessage := messageHandleObject.MQue[0]
		messageHandleObject.mu.Unlock()

		serverMessage := &chatserver.FromServer{
			Name:             currentMessage.ClientName,
			Body:             currentMessage.MessageBody,
			IsSystemMessage:  currentMessage.IsSystemMessage,
			Timestamp:        currentMessage.Timestamp,
			LamportTimestamp: LamportTimestamp,
		}

		cs.mu.Lock()
		for client := range cs.clients {
			if currentMessage.ClientUniqueCode == cs.clientMetaData[client]["clientid"][0] && !currentMessage.IsSystemMessage {
				continue
			}
			LamportTimestamp++
			serverMessage.LamportTimestamp = LamportTimestamp
			err := client.Send(serverMessage)
			if err != nil {
				log.Printf("Failed to send message to client: %v", err)
				delete(cs.clients, client)
			}
		}
		cs.mu.Unlock()

		messageHandleObject.mu.Lock()
		if len(messageHandleObject.MQue) > 0 {
			messageHandleObject.MQue = messageHandleObject.MQue[1:]
		}
		messageHandleObject.mu.Unlock()
	}
}

// main initializes and starts the gRPC server.
func main() {
	Port := os.Getenv("PORT")
	if Port == "" {
		Port = "5001"
	}

	listen, err := net.Listen("tcp", ":"+Port)
	if err != nil {
		log.Fatalf("Could not listen @ %v :: %v", Port, err)
	}
	log.Println("Server listening @ :" + Port)

	grpcserver := grpc.NewServer()
	chatServer := &ChatServer{
		clients:        make(map[chatserver.Services_ChatServiceServer]bool),
		clientMetaData: make(map[chatserver.Services_ChatServiceServer]metadata.MD),
	}

	chatserver.RegisterServicesServer(grpcserver, chatServer)

	if err := grpcserver.Serve(listen); err != nil {
		log.Fatalf("Failed to start gRPC Server :: %v", err)
	}
}