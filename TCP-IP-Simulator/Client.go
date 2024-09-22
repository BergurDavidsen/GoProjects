package main

import (
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"net"
	"os"
	"time"
)

type Packet struct {
	Syn, Ack int
	Data     string
}

func main() {
	seqNum := rand.IntN(10000)

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
	packet := Packet{Syn: seqNum}

	syn, err := toBytes(packet)

	if err != nil {
		fmt.Println(err)
		return
	}
	for i := 0; i < 3; i++ {
		fmt.Println("Client sending:", string(syn))
		_, err = conn.Write(syn)
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

		syn_ack := toPacket(buf[:n])

		if err == nil && syn_ack.Ack == seqNum+1 {
			fmt.Println("Client received:", syn_ack)

			// Send ACK

			var message string

			if len(os.Args) < 1 {
				message = "Hello World"
			} else {
				message = os.Args[1]
			}

			packet = Packet{Syn: syn_ack.Ack, Ack: syn_ack.Syn + 1, Data: message}
			ack, err := toBytes(packet)
			fmt.Println("Client sending:", string(ack))
			if err != nil {
				fmt.Println(err)
				return
			}
			_, err = conn.Write(ack)
			if err != nil {
				fmt.Println("Error writing to connection:", err)
				return
			}

			// Send "Hello world" message

			return
		}
		fmt.Println("Error reading from connection, retrying:", err)
	}
	fmt.Println("Failed to receive SYN-ACK after retries")
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
