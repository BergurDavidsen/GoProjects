package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"time"
)

// Packet represents a TCP packet with SYN, ACK, and Data fields.
type Packet struct {
	Syn, Ack int
	Data     string
}

func main() {
	// Start a TCP server listening on port 8000.
	listener, err := net.Listen("tcp", ":8000")
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	defer func(listener net.Listener) {
		err := listener.Close()
		if err != nil {
			fmt.Println("Error closing listener:", err)
		}
	}(listener)
	fmt.Println("Server is listening on port 8000")

	for {
		// Accept incoming connections.
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		go handleConnection(conn)
	}
}

// handleConnection handles the incoming connection.
func handleConnection(conn net.Conn) {
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			fmt.Println("Error closing connection:", err)
		}
	}(conn)

	// Receive SYN packet.
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Error reading from connection:", err)
		return
	}
	syn := toPacket(buf[:n])
	fmt.Println("Server received:", string(buf[:n]))

	// Generate a random sequence number for the SYN-ACK packet.
	seqNum := rand.Intn(10000)
	packet := Packet{Syn: seqNum, Ack: syn.Syn + 1}

	// Send SYN-ACK packet with retry mechanism.
	synAck, err := toBytes(packet)
	if err != nil {
		fmt.Println("Error converting packet to bytes:", err)
		return
	}
	for i := 0; i < 3; i++ {
		fmt.Println("Server sending:", string(synAck))
		_, err = conn.Write(synAck)
		if err != nil {
			fmt.Println("Error writing to connection:", err)
			return
		}

		// Wait for ACK packet with timeout.
		err := conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		if err != nil {
			fmt.Println("Error setting read deadline:", err)
			return
		}
		n, err = conn.Read(buf)
		if err == nil {
			ack := toPacket(buf[:n])
			fmt.Println("Server received:", string(buf[:n]))

			// Print the received "Hello world" message.
			fmt.Printf("Server message: %s\n", ack.Data)
			return
		}
		fmt.Println("Error reading from connection, retrying:", err)
	}
	fmt.Println("Failed to receive ACK after retries")
}

// toPacket converts a byte slice to a Packet struct.
func toPacket(buf []byte) Packet {
	var packet Packet
	err := json.Unmarshal(buf, &packet)
	if err != nil {
		fmt.Println("Error unmarshalling packet:", err)
	}
	return packet
}

// toBytes converts a Packet struct to a byte slice.
func toBytes(packet Packet) ([]byte, error) {
	return json.Marshal(packet)
}
