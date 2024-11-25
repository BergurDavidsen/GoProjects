package main

import (
	"Replication/Service"
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"google.golang.org/grpc"
)

// AuctionServer represents the gRPC server handling auction operations.
// It manages the auction state, including bids, auction duration, and item changes.
type AuctionServer struct {
	Service.UnimplementedAuctionServiceServer                                       // Embedding the unimplemented server interface
	clients                                   map[Service.AuctionServiceServer]bool // Connected clients (not currently utilized)
	mu                                        sync.Mutex                            // Mutex to synchronize access to auction state
	highestBid                                int32                                 // Current highest bid
	auctionOver                               bool                                  // Indicates if the auction is over
	startTime                                 time.Time                             // Start time of the current auction
	duration                                  time.Duration                         // Duration of the auction
	isStarted                                 bool                                  // Indicates if the auction has started
	currentWinner                             string                                // Current highest bidder
	isChangingItems                           bool
	port                                      string // Indicates if the auction items are being changed
}

var logger *log.Logger // Global logger instance

// StartAuction initializes and starts the auction for a predefined duration.
// After the auction ends, it triggers the item-changing process.
func (as *AuctionServer) StartAuction() {
	as.startTime = time.Now()
	as.isStarted = true
	as.auctionOver = false
	as.duration = 20 * time.Second
	as.highestBid = 0

	logger.Printf("Server port @%v: Auction started", as.port)

	// Wait for the auction duration to elapse
	<-time.NewTimer(as.duration).C

	as.auctionOver = true
	as.isStarted = false
	logger.Printf("Server port @%v: Auction has ended", as.port)

	// Trigger the item-changing process
	go as.ChangeAuctionItems()
}

// ChangeAuctionItems handles the transition to new auction items.
// It temporarily stops the auction for 10 seconds before resetting the state.
func (as *AuctionServer) ChangeAuctionItems() {
	as.isChangingItems = true

	logger.Printf("Server port @%v: Changing items", as.port)

	// Simulate a delay for changing items
	<-time.NewTimer(10 * time.Second).C
	logger.Printf("Server port @%v: Items changed", as.port)
	as.isChangingItems = false
}

func init() {
	// Initialize the logger to write to log.txt
	file, err := os.OpenFile("log.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
		os.Exit(1)

	}

	// Create a new logger instance
	logger = log.New(file, "", log.LstdFlags)
}

// Bid processes a bid request from a client.
// It validates the bid and updates the highest bid if the bid is successful.
func (as *AuctionServer) Bid(ctx context.Context, in *Service.BidRequest) (*Service.Ack, error) {
	as.mu.Lock()
	defer as.mu.Unlock()

	// Reject bids during item changes
	if as.isChangingItems {
		return &Service.Ack{Ack: "Exception: changing items"}, nil
	}

	// Start the auction if not already started
	if !as.isStarted {
		go as.StartAuction()
	}

	logger.Printf("Server port @%s: Bid received: %d from %s", as.port, in.Amount, in.Bidder)

	// Reject bids if the auction is over
	if as.auctionOver {
		return &Service.Ack{Ack: "Exception: auction over"}, nil
	}

	// Update the highest bid if the bid amount is greater
	if in.Amount > as.highestBid {
		as.highestBid = in.Amount
		as.currentWinner = in.Bidder
		return &Service.Ack{Ack: "Success"}, nil
	}

	// Return failure acknowledgment if the bid is not high enough
	return &Service.Ack{Ack: "Fail"}, nil
}

// Result returns the current state of the auction, including the highest bid and the highest bidder.
func (as *AuctionServer) Result(ctx context.Context, in *Service.Empty) (*Service.ResultResponse, error) {
	as.mu.Lock()
	defer as.mu.Unlock()

	return &Service.ResultResponse{
		IsOver:        as.auctionOver,
		HighestBid:    as.highestBid,
		HighestBidder: as.currentWinner,
	}, nil
}

// Start initializes and starts the gRPC server, listening on the provided port.
func (as *AuctionServer) Start() {
	// Start the server
	// Validate that a port number is provided
	if len(os.Args) != 2 {
		log.Fatalf("Please provide the port number")
		os.Exit(1)

	}

	port := os.Args[1]
	as.port = port
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Could not listen @%v :: %v", port, err)
		logger.Fatalf("Server port @%v: Could not listen @ %v :: %v", port, port, err)
	}
	log.Println("Server listening @" + port)
	logger.Println("Server listening @" + port)

	grpcServer := grpc.NewServer()

	// Register the AuctionService server
	Service.RegisterAuctionServiceServer(grpcServer, as)

	// Start serving gRPC requests
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to start gRPC Server :: %v", err)
		logger.Fatalf("Server port @%s: Failed to start gRPC Server :: %v", port, err)
	}
}

func main() {
	// Create and start the AuctionServer instance
	server := AuctionServer{
		highestBid:  0,
		auctionOver: false,
		clients:     make(map[Service.AuctionServiceServer]bool),
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	// Start a goroutine to handle signals
	go func() {
		<-signalChan // Block until a signal is received
		logger.Printf("Server @%v disconnected", os.Args[1])

		// Exit the program
		os.Exit(0)
	}()

	server.Start()
}
