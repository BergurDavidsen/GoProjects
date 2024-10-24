package main

import (
	"log"
	"net"
	"os"
	"google.golang.org/grpc"
)

func main() {
	
	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}
	
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", port, err)
	}
	log.Println("Server listening on port", port)

	grpcserver = grpc.NewServer()
	err = grpcserver.Serve(listener)
	if err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
