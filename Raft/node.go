package main

import (
	"Raft/Service"
	"fmt"
	"log"
	"net"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Node struct {
	Service.UnimplementedAuctionServiceServer
	connections []Service.AuctionServiceClient
	port        string
	isleader    bool
	mu          sync.Mutex
}

var ports = []string{"5001", "5002", "5003"}

func (n *Node) connectToPeers(ports []string) {

	for _, port := range ports {
		if port == n.port {
			continue // skip self connection
		}

		go n.connect(port)
	}
}

func (n *Node) connect(port string) {
	conn, err := grpc.NewClient(
		fmt.Sprintf("localhost:%s", port), // Server address
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	if err != nil {
		log.Printf("Retrying to connect to port :%s", port)
		n.connect(port)
	}

	// Add the connection to the client's list
	n.mu.Lock()
	defer n.mu.Unlock()
	n.connections = append(n.connections, Service.NewAuctionServiceClient(conn))
}

func listener(ports []string) (string, net.Listener, error) {
	for _, port := range ports {
		listener, err := net.Listen("tcp", "localhost:"+port)

		if err != nil {
			log.Printf("failed to listen on :%s", port)
			continue
		}
		
		return port, listener, nil
	}

	return "", nil, fmt.Errorf("no available ports to listen to")
}

func (n *Node) listen() {
	for{
		if(!n.isleader){
			continue
		}

		for  {

		}
		
	}
}

func (n *Node) broadcast() {

}



func main() {

	port, listener, err := listener(ports)
	if err != nil {
		log.Fatalf("error :: %s", err)
	}
	log.Printf("Node started listening on port: %s", port)
	node := &Node{
		isleader: false,
		port:     port,
	}

	node.connectToPeers(ports)



	defer listener.Close()

	grpcServer := grpc.NewServer()

	// Register the AuctionService server
	Service.RegisterAuctionServiceServer(grpcServer, node)

	// Start serving gRPC requests
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to start gRPC Server :: %v", err)
	}
}
