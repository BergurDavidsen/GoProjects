package main

import (
	"AuctionSystem/Service"
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	//"strconv"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	//"google.golang.org/grpc/metadata"
)

type ClientHandle struct {
	conn []Service.AuctionServiceClient
}

func reader(client *ClientHandle) {
	reader := bufio.NewReader(os.Stdin)
	
	help:
		log.Println("Enter a command:")
		log.Println("  bid <price>  - Place a bid with the given price")
		log.Println("  result        - Query the auction status")
	
	for {
		// Show available commands to the user
		

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

			bid(price, client)

		case "result":
			getResult(client)

		default:
			log.Println("Unknown command. Please try again.")
			goto help
		}
	}
}

func bid(amount int, c *ClientHandle) {
	for _, conn := range c.conn {
		ack, err := conn.Bid(context.Background(), &Service.BidRequest{
			Amount: int32(amount),
		})
		if err != nil {
			log.Printf("Error in placing bid :: %s\n", err)
			return
		}
		
		if ack == nil {
			log.Println("Error in placing bid :: Error 500")
			return
		}

		log.Printf(ack.Ack)
	}

}

func getResult(client *ClientHandle) {
	for _, conn := range client.conn {
		result, err := conn.Result(context.Background(), &Service.Empty{})
		if err != nil {
			log.Printf("Error :: %s", err)
		}
		if(result.IsOver){
			log.Printf("🔨 Auction is over. Winning bid was %d\n", result.HighestBid)
		} else {
			log.Printf("🔨 Auction is still running. Current highest bid is %d", result.HighestBid)
		}
	}
}

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s <list-server-ports>", os.Args[0:])
	}

	client := ClientHandle {}
	
	ports := os.Args[1:]

	for _, port := range ports {
		
		conn, err := grpc.NewClient(
			fmt.Sprintf("localhost:%s", port),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)

		if err != nil {
			log.Fatalf("Error to make connection :: %s", err)
		}

		client.conn = append(client.conn, Service.NewAuctionServiceClient(conn))
	}

	go reader(&client)
	select {}
}
