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

var LamportTimestamp uint32 = 1

// ClientHandle
type ClientHandle struct {
	stream     chatserver.Services_ChatServiceClient
	clientName string
	clientId   int
}

// send message
func (ch *ClientHandle) sendMessage() {

	// create a loop
	for {

		reader := bufio.NewReader(os.Stdin)
		clientMessage, err := reader.ReadString('\n')
		if err != nil {
			log.Fatalf(" Failed to read from console :: %v", err)
		}
		clientMessage = strings.Trim(clientMessage, "\r\n")

		LamportTimestamp++

		clientMessageBox := &chatserver.FromClient{
			Name:             ch.clientName,
			Body:             clientMessage,
			LamportTimestamp: LamportTimestamp,
		}

		err = ch.stream.Send(clientMessageBox)

		if err != nil {
			log.Printf("Error while sending message to server :: %v", err)
		}

		fmt.Printf("Your message was sent\n")

	}

}

//receive message

func (ch *ClientHandle) receiveMessage() {
	for {
		mssg, err := ch.stream.Recv()
		if err != nil {
			log.Printf("Error in receiving message from server :: %v", err)
			continue
		}
		// Display messages with timestamps
		LamportTimestamp = (max(mssg.LamportTimestamp, LamportTimestamp) + 1)
		if mssg.IsSystemMessage {
			fmt.Printf("[%s] {%d}\n🔔 System: %s\n", mssg.Timestamp, LamportTimestamp, mssg.Body)
		} else {
			fmt.Printf("\n[%s] {%d}\n%s:%s\n", mssg.Timestamp, LamportTimestamp, mssg.Name, mssg.Body)
		}

	}
}

/*
func (ch *ClientHandle) join(){
	LamportTimestamp++

	clientMessageBox := &chatserver.FromClient{
		Name: ch.clientName,
		LamportTimestamp: LamportTimestamp,

	}

	err := ch.stream.Send(clientMessageBox)

	if err != nil {
		log.Printf("Error while sending message to server :: %v", err)
	}
}
*/

func (ch *ClientHandle) clientConfig() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("Your Name : ")
	name, err := reader.ReadString('\n')
	if err != nil {
		log.Fatalf(" Failed to read from console :: %v", err)
	}
	ch.clientName = strings.Trim(name, "\r\n")
	ch.clientId = rand.Intn(1e6)

}

func main() {
	// configure client
	ch := ClientHandle{}
	ch.clientConfig()

	// enter localhost
	fmt.Println("Enter Server IP:Port ::: ")
	reader := bufio.NewReader(os.Stdin)
	serverID, err := reader.ReadString('\n')
	if err != nil {
		log.Printf("Failed to read from console :: %v", err)
	}
	serverID = strings.Trim(serverID, "\r\n")
	log.Println("Connecting : " + serverID)

	//connect to grpc server
	conn, err := grpc.NewClient(serverID, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Faile to conncet to gRPC server :: %v", err)
	}
	defer conn.Close()
	LamportTimestamp++

	//call ChatService to create a stream
	client := chatserver.NewServicesClient(conn)

	// add metadata to the context
	md := metadata.Pairs("clientId", strconv.Itoa(ch.clientId),
		"clientName", ch.clientName, "Lamport", strconv.FormatUint(uint64(LamportTimestamp), 10))
	ctx := metadata.NewOutgoingContext(context.Background(), md)

	// add stream to ClientHandle
	stream, err := client.ChatService(ctx)
	if err != nil {
		log.Fatalf("Failed to call ChatService :: %v", err)
		return
	}
	ch.stream = stream
	// implement communication with gRPC server
	go ch.sendMessage()
	go ch.receiveMessage()

	//blocker
	bl := make(chan bool)
	<-bl

}
