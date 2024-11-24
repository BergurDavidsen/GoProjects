package main

import (
	"Replication/Service"
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// ClientHandle represents the client interacting with multiple auction services.
// It holds the gRPC connections and the client's name.
type ClientHandle struct {
	conn []Service.AuctionServiceClient // List of auction service clients
	name string                         // Name of the client (bidder)
}

// reader handles user input and executes the corresponding auction commands.
// Commands:
//   - `bid <price>`: Places a bid of the specified amount.
//   - `result`: Retrieves the current status of the auction.
func reader(client *ClientHandle) {
	reader := bufio.NewReader(os.Stdin)

help: // Label for displaying the help message when an invalid command is entered
	log.Println("Enter a command:")
	log.Println("  bid <price>  - Place a bid with the given price")
	log.Println("  result        - Query the auction status")

	for {
		fmt.Print("> ")
		// Read user input
		input, err := reader.ReadString('\n')
		if err != nil {
			log.Println("Error reading input:", err)
			continue
		}

		// Trim and split the input into a command and arguments
		input = strings.TrimSpace(input)
		message := strings.Split(strings.ToLower(input), " ")

		// Validate input
		if len(message) == 0 {
			log.Println("Invalid input. Please try again.")
			continue
		}

		// Process the command
		switch message[0] {
		case "bid":
			// Handle the "bid" command
			if len(message) != 2 {
				log.Println("Invalid format. Usage: bid <price>")
				continue
			}

			// Convert the bid price to an integer
			price, err := strconv.Atoi(message[1])
			if err != nil {
				log.Println("Invalid price. Please enter a numeric value.")
				continue
			}

			// Place the bid
			bid(price, client)

		case "result":
			// Handle the "result" command
			getResult(client)

		default:
			// Unknown command: display help and re-enter the loop
			log.Println("Unknown command. Please try again.")
			goto help
		}
	}
}

// bid sends a bid request to all connected auction servers.
// The function logs the acknowledgment received from the first responding server.
func bid(amount int, c *ClientHandle) {
	outputArray := make([]string, 0)

	// Iterate over all connections and send the bid
	for _, conn := range c.conn {
		ack, err := conn.Bid(context.Background(), &Service.BidRequest{
			Amount: int32(amount), // Bid amount
			Bidder: c.name,        // Bidder's name
		})

		if err != nil {
			continue // Skip this server on error
		}

		if ack == nil {
			continue // Skip if no acknowledgment is received
		}

		outputArray = append(outputArray, ack.Ack)
	}

	// Log the acknowledgment from the first server that responds
	if len(outputArray) > 0 {
		log.Println(outputArray[0])
	} else {
		log.Println("No acknowledgment received from servers.")
	}
}

// getResult queries the current auction status from all connected servers.
// The status includes whether the auction is over and the current highest bid.
func getResult(client *ClientHandle) {
	for _, conn := range client.conn {
		result, err := conn.Result(context.Background(), &Service.Empty{})
		if err != nil {
			log.Printf("Error :: %s", err)
			continue
		}

		// Log the auction status
		if result.IsOver {
			log.Printf("🔨 Auction is over. Winning bid was %d and was from %s\n", result.HighestBid, result.HighestBidder)
		} else {
			log.Printf("🔨 Auction is still running. Current highest bid is %d was from %s\n", result.HighestBid, result.HighestBidder)
		}
	}
}

func main() {
	// Ensure correct usage with required arguments
	if len(os.Args) < 3 {
		log.Fatalf("Usage: %s <name> <list-server-ports>", os.Args[0:]) // Display usage and exit
	}

	// Create a new client handle
	client := ClientHandle{}
	name := os.Args[1]  // Client's name
	ports := os.Args[2:] // List of server ports

	client.name = name

	// Establish connections to all auction servers
	for _, port := range ports {

		conn, err := grpc.NewClient(
			fmt.Sprintf("localhost:%s", port), // Server address
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)

		if err != nil {
			log.Fatalf("Error to make connection :: %s", err) // Exit on connection failure
		}

		// Add the connection to the client's list
		client.conn = append(client.conn, Service.NewAuctionServiceClient(conn))
	}

	// Start reading commands from the user
	reader(&client)
}
