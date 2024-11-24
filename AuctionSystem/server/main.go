package main

import (
	"AuctionSystem/Service"
	"context"
	"log"
	"net"
	"os"
	"sync"
	"time"

	"google.golang.org/grpc"
)

type AuctionServer struct {
	Service.UnimplementedAuctionServiceServer
	clients          map[Service.AuctionServiceServer]bool
	mu               sync.Mutex
    highestBid int32
	auctionOver bool
	startTime time.Time
	duration time.Duration
	isStarted bool
	isChangingItems bool
}


func (as *AuctionServer) StartAuction() {
	as.startTime = time.Now()
    as.isStarted = true

	log.Println("Auction started")

	<- time.NewTimer(as.duration).C
	
	as.auctionOver = true
    as.isStarted = false
    log.Println("Auction has ended")
	
	go as.ChangeAuctionItems()
	
}



func (as *AuctionServer) ChangeAuctionItems() {
	as.isChangingItems = true
	as.auctionOver = true
	as.highestBid = 0

	
	log.Println("Changing items")
	<- time.NewTimer(10*time.Second).C
	log.Println("Changing items")

	as.isChangingItems = false
	as.auctionOver = false
}

func (as *AuctionServer) Bid(ctx context.Context, in *Service.BidRequest) (*Service.Ack, error) {
	as.mu.Lock()
	defer as.mu.Unlock()
	if as.isChangingItems {
		return &Service.Ack{Ack: "Exception: changing items"}, nil
	}
	if !as.isStarted {
		go as.StartAuction()
	}
	
	log.Printf("Bid received on: %d\n", in.Amount)

	if as.auctionOver {
		return &Service.Ack{Ack: "Exception: auction over"}, nil
	} else {
		if in.Amount > as.highestBid {
			as.highestBid = in.Amount
			return &Service.Ack{Ack: "Success"}, nil
		} else {
			return &Service.Ack{Ack: "Fail"}, nil
		}
    }	
}

func (as *AuctionServer) Result(ctx context.Context, in *Service.Empty) (*Service.ResultResponse, error) {
	as.mu.Lock()
	defer as.mu.Unlock()
	return &Service.ResultResponse{IsOver: as.auctionOver, HighestBid: as.highestBid}, nil

}

func (as *AuctionServer) Start() {
	// Start the server
	if(len(os.Args) != 2) {
		log.Fatalf("Please provide the port number")
	}
	port := os.Args[1]
	listener, err := net.Listen("tcp", ":" + port)
	if err != nil {
		log.Fatalf("Could not listen @ %v :: %v", port, err)
	}
	log.Println("Server listening @ :" + port)

	grpcServer := grpc.NewServer()
	auctionServer := AuctionServer{
		clients:          make(map[Service.AuctionServiceServer]bool),
		
	}

	Service.RegisterAuctionServiceServer(grpcServer, &auctionServer)

	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to start gRPC Server :: %v", err)
	}
}

func main() {
	server := AuctionServer{
		highestBid: 0,
		auctionOver: false,
		clients: make(map[Service.AuctionServiceServer]bool),
		duration: 20 * time.Second,
	}

	server.Start()
}
