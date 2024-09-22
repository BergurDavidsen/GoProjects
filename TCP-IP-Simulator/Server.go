package main

import (
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"net"
	"time"
)

type Packet struct {
	Syn, Ack int
	Data     string
}

func main() {
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
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			fmt.Println("Error closing connection:", err)
		}
	}(conn)

	// Receive SYN
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)

	if err != nil {
		fmt.Println("Error reading from connection:", err)
		return
	}
	fmt.Println("Server received:", string(buf[:n]))

	//this is to map the packets to different fields
	syn := toPacket(buf[:n])

	seqNum := rand.IntN(10000)

	packet := Packet{Syn: seqNum, Ack: syn.Syn + 1}
	// Send SYN-ACK with retry mechanism
	syn_ack, err := toBytes(packet)
	if err != nil {
		fmt.Println(err)
		return
	}
	for i := 0; i < 3; i++ {
		fmt.Println("Server sending:", string(syn_ack))

		_, err = conn.Write(syn_ack)
		if err != nil {
			fmt.Println("Error writing to connection:", err)
			return
		}

		// Wait for ACK with timeout
		err := conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		if err != nil {
			fmt.Println("Error setting read deadline:", err)
			return
		}
		n, err = conn.Read(buf)

		ack := toPacket(buf)

		if err == nil {
			fmt.Println("Server received:", ack)

			// Receive "Hello world" message
			fmt.Printf("Server message: %v", ack.Data)
		}
		fmt.Println("Error reading from connection, retrying:", err)
	}
	fmt.Println("Failed to receive ACK after retries")
}

func toPacket(buf []byte) Packet {

	var jsonMap Packet
	json.Unmarshal(buf, &jsonMap)

	return jsonMap
}

func toBytes(packet Packet) ([]byte, error) {
	var bytes []byte

	bytes, err := json.Marshal(packet)

	return bytes, err
}
