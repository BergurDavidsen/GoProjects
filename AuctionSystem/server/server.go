package main

import (
	"AuctionSystem/Service"
	"fmt"
	"log"
	"net"
	"sync"

	"google.golang.org/grpc"
)

var ports = []string{"5001", "5002", "5003", "5004", "5005", "5006", "5007", "5008", "5009", "5010"}

type AuctionServer struct {
	Service.UnimplementedAuctionServiceServer
	clients 	map[Service.AuctionServiceServer]bool
	mu 			sync.Mutex
}

func (as AuctionServer) AuctionService(csi Service.AuctionService_AuctionServiceServer) error {
	errch := make(chan error)
	as.mu.Lock()

	md := 

	as.mu.Unlock()
	return <-errch
}

func listener() (*net.Listener, string, error) {
	for _, port := range ports {
		listener, err := net.Listen("tcp", fmt.Sprintf(":%v", port))
		if err != nil {
			log.Printf("Could not listen @ %v :: %v", port, err)
			continue
		}
		return &listener, port, nil
	}
	return nil, "err",  fmt.Errorf("no available ports in the list")
}


func main() {
	listener, port, err := listener()
	if err != nil {
		log.Fatalf("Could not listen @ %v :: %v", port, err)
	}
	log.Println("Server listening @ :" + port)

	grpcServer := grpc.NewServer()
	auctionServer := AuctionServer{
		clients: make(map[Service.AuctionServiceServer]bool),
	}

	Service.RegisterAuctionServiceServer(grpcServer, auctionServer)

	if err := grpcServer.Serve(*listener); err != nil {
		log.Fatalf("Failed to start gRPC Server :: %v", err)
	}

	
}