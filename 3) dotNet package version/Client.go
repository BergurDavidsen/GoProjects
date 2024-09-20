package main

import (
	"fmt"
	"net"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		return
	}
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {

		}
	}(conn)

	// Send SYN
	syn := "SYN"
	fmt.Println("Client sending:", syn)
	_, err = conn.Write([]byte(syn))
	if err != nil {
		fmt.Println("Error writing to connection:", err)
		return
	}

	// Receive SYN-ACK
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Error reading from connection:", err)
		return
	}
	fmt.Println("Client received:", string(buf[:n]))

	// Send ACK
	ack := "ACK"
	fmt.Println("Client sending:", ack)
	_, err = conn.Write([]byte(ack))
	if err != nil {
		fmt.Println("Error writing to connection:", err)
		return
	}
}
