package main

import (
	"context"
	"log"
	"net"
	"os"
	"strconv"
	"time"

	"math/rand"

	pb "github.com/BergurDavidsen/GoProjects/Consensus/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type (
	Node struct {
		nextNode  pb.TokenServiceClient
		id        int
		port      string
		neighbors []string
	}

	TokenServer struct {
		pb.UnimplementedTokenServiceServer
		node *Node
	}
)

func (t *TokenServer) PassToken(ctx context.Context, token *pb.Token) (*pb.Ack, error) {
	log.Printf("Node %d received token: %s from holder: %s", t.node.id, token.Token, token.Holder)

	// enter critical section if wanted
	EnterCriticalSection()

	// send token to next node
	sendToken(t.node)
	return &pb.Ack{Success: true}, nil
}

func sendToken(node *Node) {
	//consruct token
	newToken := &pb.Token{
		Message: "hello",
		Id:      int32(rand.Uint32()),
		Token:   "Authentication token",
		Holder:  strconv.Itoa(node.id),
	}

	_, err := node.nextNode.PassToken(context.Background(), newToken)
	if err != nil {
		log.Printf("Error sending token: %v", err)
	} else {
		log.Printf("Node %d sent token to node %d\n", node.id, node.id+1)
	}

}

func EnterCriticalSection() {
	log.Printf("Entering critical section\n")
	time.Sleep(2 * time.Second)
	log.Printf("Exiting critical section\n")
}

func main() {
	//promt user for port and node id
	var node = &Node{}

	if len(os.Args) < 4 {
		panic("Usage: go run node.go <port> <id> <neighbors>")
	}

	node.id, _ = strconv.Atoi(os.Args[1])
	node.port = os.Args[2]
	node.neighbors = os.Args[3:]

	tokenServer := &TokenServer{node: node}

	//register listener
	lis, err := net.Listen("tcp", ":"+node.port)
	if err != nil {
		log.Printf("There was an error when starting listener\n")
		panic(err)
	}

	// start client and grpc server
	s := grpc.NewServer()

	nextNodeAddress := node.neighbors[0]
	conn, err := grpc.NewClient(nextNodeAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("There was an error when starting client\n")
		panic(err)
	}
	pb.RegisterTokenServiceServer(s, tokenServer)
	pb.NewTokenServiceClient(conn)
	node.nextNode = pb.NewTokenServiceClient(conn)
	go func() {
		log.Printf("Node %d listening on port %s\n", node.id, node.port)
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Failed to serve gRPC server on port %s: %v", node.port, err)
		}
	}()

	// Start by sending the token if this is the first node
	if node.id == 1 {
		log.Printf("Node %d is initiating the token passing\n", node.id)
		sendToken(node)
	}

	// Wait for user input to exit
	bl := make(chan bool)
	<-bl
}
