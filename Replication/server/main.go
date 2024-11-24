package main

import (
	"Replication/Service"
	"context"
	"log"
	"net"
	"os"
	"sync"
	"time"

	"google.golang.org/grpc"
)

// AuctionServer represents the gRPC server handling auction operations.
// It manages the auction state, including bids, auction duration, and item changes.
type AuctionServer struct {
	Service.UnimplementedAuctionServiceServer // Embedding the unimplemented server interface
	clients          map[Service.AuctionServiceServer]bool // Connected clients (not currently utilized)
	mu               sync.Mutex                           // Mutex to synchronize access to auction state
	highestBid       int32                                // Current highest bid
	auctionOver      bool                                 // Indicates if the auction is over
	startTime        time.Time                            // Start time of the current auction
	duration         time.Duration                        // Duration of the auction
	isStarted        bool                                 // Indicates if the auction has started
	currentWinner    string                               // Current highest bidder
	isChangingItems  bool                                 // Indicates if the auction items are being changed
}

// StartAuction initializes and starts the auction for a predefined duration.
// After the auction ends, it triggers the item-changing process.
func (as *AuctionServer) StartAuction() {
	as.startTime = time.Now()
	as.isStarted = true
	as.duration = 20 * time.Second

	log.Println("Auction started")

	// Wait for the auction duration to elapse
	<-time.NewTimer(as.duration).C

	as.auctionOver = true
	as.isStarted = false
	log.Println("Auction has ended")

	// Trigger the item-changing process
	go as.ChangeAuctionItems()
}

// ChangeAuctionItems handles the transition to new auction items.
// It temporarily stops the auction for 10 seconds before resetting the state.
func (as *AuctionServer) ChangeAuctionItems() {
	as.isChangingItems = true
	as.auctionOver = true

	log.Println("Changing items")

	// Simulate a delay for changing items
	<-time.NewTimer(10 * time.Second).C
	log.Println("Items changed")
	as.highestBid = 0
	as.isChangingItems = false
	as.auctionOver = false
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

	log.Printf("Bid received: %d\n", in.Amount)

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
		IsOver:       as.auctionOver,
		HighestBid:   as.highestBid,
		HighestBidder: as.currentWinner,
	}, nil
}

// Start initializes and starts the gRPC server, listening on the provided port.
func (as *AuctionServer) Start() {
	// Validate that a port number is provided
	if len(os.Args) != 2 {
		log.Fatalf("Please provide the port number")
	}

	port := os.Args[1]
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Could not listen @ %v :: %v", port, err)
	}
	log.Println("Server listening @ :" + port)

	grpcServer := grpc.NewServer()
	auctionServer := AuctionServer{
		clients: make(map[Service.AuctionServiceServer]bool),
	}

	// Register the AuctionService server
	Service.RegisterAuctionServiceServer(grpcServer, &auctionServer)

	// Start serving gRPC requests
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to start gRPC Server :: %v", err)
	}
}

func main() {
	// Create and start the AuctionServer instance
	server := AuctionServer{
		highestBid:  0,
		auctionOver: false,
		clients:     make(map[Service.AuctionServiceServer]bool),
	}

	server.Start()
}
