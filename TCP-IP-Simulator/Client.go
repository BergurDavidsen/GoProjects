package main

import (
	"fmt"
	"net"
	"time"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:8000")
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		return
	}
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			fmt.Println("Error closing connection:", err)
		}
	}(conn)

	// Send SYN with retry mechanism
	syn := "SYN"
	for i := 0; i < 3; i++ {
		fmt.Println("Client sending:", syn)
		_, err = conn.Write([]byte(syn))
		if err != nil {
			fmt.Println("Error writing to connection:", err)
			return
		}

		// Receive SYN-ACK with timeout
		buf := make([]byte, 1024)
		err := conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		if err != nil {
			fmt.Println("Error setting read deadline:", err)
			return
		}
		n, err := conn.Read(buf)
		if err == nil {
			fmt.Println("Client received:", string(buf[:n]))

			// Send ACK
			ack := "ACK"
			fmt.Println("Client sending:", ack)
			_, err = conn.Write([]byte(ack))
			if err != nil {
				fmt.Println("Error writing to connection:", err)
				return
			}

			// Send "Hello world" message
			helloMsg := "Hello world"
			fmt.Println("Client sending:", helloMsg)
			_, err = conn.Write([]byte(helloMsg))
			if err != nil {
				fmt.Println("Error writing to connection:", err)
				return
			}
			return
		}
		fmt.Println("Error reading from connection, retrying:", err)
	}
	fmt.Println("Failed to receive SYN-ACK after retries")
}
