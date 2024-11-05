package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	rpc "github.com/BergurDavidsen/GoProjects/Consensus/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Node struct {
	rpc.UnimplementedNodeServiceServer
	id              int
	peers           []rpc.NodeServiceClient
	peerAddresses   []string
	mu              sync.Mutex
	timestamp       int64
	inCritical      bool
	requestQueue    []int
	receivedReplies map[int]bool
}

func NewNode(id int, peerAddresses []string) *Node {
	return &Node{
		id:              id,
		peerAddresses:   peerAddresses,
		receivedReplies: make(map[int]bool),
	}
}

// gRPC handler for access requests
func (node *Node) RequestAccess(ctx context.Context, req *rpc.RequestMessage) (*rpc.ResponseMessage, error) {
	node.mu.Lock()
	defer node.mu.Unlock()

	// Reply to request if we are not in CS or have lower-priority request
	if !node.inCritical || (node.timestamp == 0) || (req.Timestamp < node.timestamp || (req.Timestamp == node.timestamp && int(req.NodeId) < node.id)) {
		return &rpc.ResponseMessage{PermissionGranted: true}, nil
	}

	// Otherwise, queue the request and respond later
	node.requestQueue = append(node.requestQueue, int(req.NodeId))
	return &rpc.ResponseMessage{PermissionGranted: false}, nil
}

// Method to enter Critical Section
func (node *Node) RequestCriticalSection() {
	node.mu.Lock()
	node.timestamp = time.Now().UnixNano()
	node.inCritical = true
	node.mu.Unlock()

	// Broadcast request to all peers
	for _, peer := range node.peers {
		go func(peer rpc.NodeServiceClient) {
			peer.RequestAccess(context.Background(), &rpc.RequestMessage{NodeId: int32(node.id), Timestamp: node.timestamp})
		}(peer)
	}

	// Wait for replies from all peers
	for len(node.receivedReplies) < len(node.peers) {
		time.Sleep(100 * time.Millisecond)
	}

	node.mu.Lock()
	fmt.Printf("Node %d: Entering Critical Section\n", node.id)
	node.mu.Unlock()
	time.Sleep(2 * time.Second) // Simulating work in Critical Section
	fmt.Printf("Node %d: Leaving Critical Section\n", node.id)

	// Reset state and respond to pending requests
	node.mu.Lock()
	node.inCritical = false
	node.timestamp = 0
	node.receivedReplies = make(map[int]bool)
	node.mu.Unlock()

	for _, queuedNode := range node.requestQueue {
		fmt.Println("Sending reply to queued node", queuedNode)
		// Send reply back to queued nodes
		// Assume `sendReply` is defined to send responses
	}
}

// Connect to other nodes
func (node *Node) ConnectToPeers() {
	for _, addr := range node.peerAddresses {
		conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatalf("Failed to connect to peer %s: %v", addr, err)
		}
		client := rpc.NewNodeServiceClient(conn)
		node.peers = append(node.peers, client)
	}
}

// Start gRPC server
func (node *Node) StartServer(port string) {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", port, err)
	}

	server := grpc.NewServer()
	rpc.RegisterNodeServiceServer(server, node)

	log.Printf("Node %d listening on port %s", node.id, port)
	if err := server.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run node.go <nodeId> <port> <peerAddress1> <peerAddress2> ...")
		return
	}

	nodeId, _ := strconv.Atoi(os.Args[1])
	port := os.Args[2]
	peerAddresses := os.Args[3:]

	node := NewNode(nodeId, peerAddresses)
	go node.StartServer(port)
	time.Sleep(1 * time.Second)
	node.ConnectToPeers()

	// Simulate node requesting CS
	for {
		time.Sleep(time.Duration(5+nodeId) * time.Second)
		node.RequestCriticalSection()
	}
}
