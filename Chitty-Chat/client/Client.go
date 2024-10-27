package main

import (
	"bufio"
	"context"
	"fmt"
	"grpcChatServer/chatserver"
	"log"
	"os"
	"strings"
	"math/rand"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var LamportTimestamp uint32 = 0



// ClientHandle
type ClientHandle struct {
	stream     chatserver.Services_ChatServiceClient
	config 	   Config
}

type Config struct {
	clientName string
	clientId int
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
			Name: ch.config.clientName,
			Body: clientMessage,
			LamportTimestamp: LamportTimestamp,
		}

		err = ch.stream.Send(clientMessageBox)

		if err != nil {
			log.Printf("Error while sending message to server :: %v", err)
		}

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
		
		LamportTimestamp = (max(mssg.LamportTimestamp, LamportTimestamp) + 1)

		// Display messages with timestamps
		if mssg.IsSystemMessage {
			fmt.Printf("{%d} [%s] 🔔 System: %s\n", LamportTimestamp, mssg.Timestamp, mssg.Body)
		} else {
			fmt.Printf("{%d} [%s] %s: %s\n", LamportTimestamp, mssg.Timestamp, mssg.Name, mssg.Body)
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

func (ch *ClientHandle) sync(){

}

func (ch *ClientHandle) clientConfig() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("Your Name : ")
	name, err := reader.ReadString('\n')
	if err != nil {
		log.Fatalf(" Failed to read from console :: %v", err)
	}
	ch.config.clientName = strings.Trim(name, "\r\n")
	ch.config.clientId = rand.Intn(1e6)

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

	//call ChatService to create a stream
	client := chatserver.NewServicesClient(conn)
	
	ctx := context.WithValue(context.Background(), "config", ch.config)

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