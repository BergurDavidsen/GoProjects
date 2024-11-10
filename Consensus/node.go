package main

import (
	"log"
	"net"
	"os"
	"strconv"

	pb "github.com/BergurDavidsen/GoProjects/Consensus/proto"
	"google.golang.org/grpc"
)

type Node struct {
	id        int
	port      int
	neighbors []string
}

type tokenServer struct {
	pb.UnimplementedTokenServiceServer
	nextProcessAddress string // Address of the next process in the ring
}

func listenForToken() {

}

func main() {
	//promt user for port and node id
	node := Node{}

	if len(os.Args) < 4 {
		panic("Usage: go run node.go <port> <id> <neighbors>")
	}

	node.port, _ = strconv.Atoi(os.Args[1])
	node.id, _ = strconv.Atoi(os.Args[2])
	node.neighbors = os.Args[3:]

	//register listener
	lis, err := net.Listen("tcp", ":"+strconv.Itoa(node.port))
	if err != nil {
		log.Printf("There was an error when starting listener\n")
		panic(err)
	}
	// start client and grpc server
	s := grpc.NewServer()
	pb.RegisterTokenServiceServer(s, &tokenServer{})
	s.Serve(lis)

}
