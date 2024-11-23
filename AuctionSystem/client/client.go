package main

import (
	"AuctionSystem/Service"
	"bufio"
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"

	//"strconv"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	//"google.golang.org/grpc/metadata"
)

type (
	ClientHandle struct{
		stream Service.AuctionService_AuctionServiceClient
		userName string
		id int
	}
)
 
func (c *ClientHandle) sendToStream(reader *bufio.Reader) {
	for {
		// Show available commands to the user
		log.Println("Enter a command:")
		log.Println("  bid <price>  - Place a bid with the given price")
		log.Println("  query        - Query the auction status")

		// Read user input
		input, err := reader.ReadString('\n')
		if err != nil {
			log.Println("Error reading input:", err)
			continue
		}

		// Process input: split by spaces and trim whitespace
		input = strings.TrimSpace(input)
		message := strings.Split(strings.ToLower(input), " ")

		// Validate input
		if len(message) == 0 {
			log.Println("Invalid input. Please try again.")
			continue
		}

		var request *Service.ClientRequest

		// Handle commands
		switch message[0] {
		case "bid":
			// Ensure correct number of arguments
			if len(message) != 2 {
				log.Println("Invalid format. Usage: bid <price>")
				continue
			}

			// Convert price to integer
			price, err := strconv.Atoi(message[1])
			if err != nil {
				log.Println("Invalid price. Please enter a numeric value.")
				continue
			}

			// Create bid request
			request = &Service.ClientRequest{
				Request: &Service.ClientRequest_Bid{
					Bid: &Service.BidRequest{
						Price: int32(price),
					},
				},
			}

		case "query":
			// Ensure no extra arguments
			if len(message) != 1 {
				log.Println("Invalid format. Usage: query")
				continue
			}

			// Create query request
			request = &Service.ClientRequest{
				Request: &Service.ClientRequest_Query{
					Query: &Service.QueryRequest{
						Request: true,
					},
				},
			}

		default:
			log.Println("Unknown command. Please try again.")
			continue
		}

		// Send the request to the stream
		if err := c.stream.Send(request); err != nil {
			log.Println("Failed to send request:", err)
		} else {
			log.Println("Request sent:", request)
		}
	}
}

func (c *ClientHandle) receiveFromStream() {
	// not implemented
}


func clientConfig(config string) *ClientHandle {
	return &ClientHandle{
		userName: config,
		id: rand.Intn(1e6),
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

	conn, err := grpc.NewClient(
		fmt.Sprintf("localhost:%s", Tinput[1]), 
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	if err != nil {
		log.Fatalf("Failed to connect to gRPC server :: %v", err)
	}
	defer conn.Close()

	var clientService = Service.NewAuctionServiceClient(conn)
	
	/* doesnt work
	ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs(
		"clientId", strconv.Itoa(client.id), 
		"clientName", client.userName,
	))*/

	

	stream, err := clientService.AuctionService(context.Background())
	if err != nil {
		log.Fatalf("Failed to call ChatService :: %v", err)
		return
	}

	log.Println("Successfully connected to Auction house")
	client.stream = stream

	go client.sendToStream(reader)
	go client.receiveFromStream()

	select {}
}