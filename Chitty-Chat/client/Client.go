package main

import (
	"bufio"
	"context"
	"fmt"
	"grpcChatServer/chatserver"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

var LamportTimestamp uint32 = 1 // Lamport clock for message ordering

// ClientHandle manages client properties and stream communication
type ClientHandle struct {
	stream     chatserver.Services_ChatServiceClient
	clientName string
	clientId   int
}

// sendMessage reads user input and sends messages to the server
func (ch *ClientHandle) sendMessage() {
	for {
		reader := bufio.NewReader(os.Stdin)
		clientMessage, err := reader.ReadString('\n')
		if err != nil {
			log.Fatalf("Failed to read from console :: %v", err)
		}
		clientMessage = strings.Trim(clientMessage, "\r\n")

		LamportTimestamp++
		clientMessageBox := &chatserver.FromClient{
			Name:             ch.clientName,
			Body:             clientMessage,
			LamportTimestamp: LamportTimestamp,
		}

		if err = ch.stream.Send(clientMessageBox); err != nil {
			log.Printf("Error sending message to server :: %v", err)
		}

		fmt.Printf("Your message was sent\n")
	}
}

// receiveMessage listens for messages from the server
func (ch *ClientHandle) receiveMessage() {
	for {
		mssg, err := ch.stream.Recv()
		if err != nil {
			log.Printf("Error receiving message from server :: %v", err)
			continue
		}
		LamportTimestamp = max(mssg.LamportTimestamp, LamportTimestamp) + 1
		if mssg.IsSystemMessage {
			fmt.Printf("[%s] {%d}\n🔔 System: %s\n", mssg.Timestamp, LamportTimestamp, mssg.Body)
		} else {
			fmt.Printf("\n[%s] {%d}\n%s: %s\n", mssg.Timestamp, LamportTimestamp, mssg.Name, mssg.Body)
		}
	}
}

// clientConfig sets up the client name and ID
func (ch *ClientHandle) clientConfig() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("Your Name: ")
	name, err := reader.ReadString('\n')
	if err != nil {
		log.Fatalf("Failed to read from console :: %v", err)
	}
	ch.clientName = strings.Trim(name, "\r\n")
	ch.clientId = rand.Intn(1e6)
}

func main() {
	ch := ClientHandle{}
	ch.clientConfig()

	fmt.Println("Enter Server IP:Port ::: ")
	reader := bufio.NewReader(os.Stdin)
	serverID, err := reader.ReadString('\n')
	if err != nil {
		log.Printf("Failed to read from console :: %v", err)
	}
	serverID = strings.Trim(serverID, "\r\n")
	log.Println("Connecting to " + serverID)

	conn, err := grpc.Dial(serverID, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to gRPC server :: %v", err)
	}
	defer conn.Close()
	LamportTimestamp++

	client := chatserver.NewServicesClient(conn)
	md := metadata.Pairs("clientId", strconv.Itoa(ch.clientId), "clientName", ch.clientName, "Lamport", strconv.FormatUint(uint64(LamportTimestamp), 10))
	ctx := metadata.NewOutgoingContext(context.Background(), md)

	stream, err := client.ChatService(ctx)
	if err != nil {
		log.Fatalf("Failed to call ChatService :: %v", err)
		return
	}
	ch.stream = stream

	go ch.sendMessage()
	go ch.receiveMessage()

	bl := make(chan bool)
	<-bl
}
