package main

import (
	"AuctionSystem/Service"
	"bufio"
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/internal/metadata"
)

type (
	ClientHandle struct{
		stream Service.AuctionService_AuctionServiceClient
		userName string
		id uint32
	}
)

func sendToStream(client *ClientHandle, reader *bufio.Reader){
	for{
		message, _ := reader.ReadString('\n')

		log.Println(message)
	}

}

func receiveFromStream(client *ClientHandle){
	// not implemented
}


func clientConfig(config string) *ClientHandle {
	return &ClientHandle{
		userName: config,
		id: rand.Uint32(),
	}
}

func main() {
	var reader = bufio.NewReader(os.Stdin)

	log.Printf("Please enter your name and port:")
	var configs, err = reader.ReadString('\n')
	if err != nil {
		log.Printf("Failed to read from console :: %v", err)
	}

	var input = strings.Trim(configs,"\r\n")

	var Tinput = strings.Split(input, " ")
	client := clientConfig(Tinput[0])

	conn, err := grpc.NewClient(fmt.Sprintf("localhost:%s", Tinput[1]), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to gRPC server :: %v", err)
	}
	defer conn.Close()

	var clientService = Service.NewAuctionServiceClient(conn)
	ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs(
		"clientId", client.id, 
		"clientName", client.userName,
	))



	
	stream, err := clientService.AuctionService(ctx)
	if err != nil {
		log.Fatalf("Failed to call ChatService :: %v", err)
		return
	}

	client.stream = stream

	go sendToStream(client, reader)
	go receiveFromStream(client)

	select {}
}