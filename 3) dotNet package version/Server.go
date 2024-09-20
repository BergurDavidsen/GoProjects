package main

import (
	"fmt"
	"net"
)

func main() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	defer listener.Close()
	fmt.Println("Server is listening on port 8080")

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
	defer conn.Close()

	// Receive SYN
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Error reading from connection:", err)
		return
	}
	fmt.Println("Server received:", string(buf[:n]))

	// Send SYN-ACK
	synAck := "SYN-ACK"
	fmt.Println("Server sending:", synAck)
	_, err = conn.Write([]byte(synAck))
	if err != nil {
		fmt.Println("Error writing to connection:", err)
		return
	}

	// Receive ACK
	n, err = conn.Read(buf)
	if err != nil {
		fmt.Println("Error reading from connection:", err)
		return
	}
	fmt.Println("Server received:", string(buf[:n]))
}
