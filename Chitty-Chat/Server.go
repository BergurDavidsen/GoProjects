package main

import (
	"google.golang.org/grpc"
)

func main() {
	s := grpc.NewServer()

	pb.RegisterGreeterServer(s, &server{})
}
