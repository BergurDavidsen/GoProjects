package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"os"
	"time"
)

// Packet represents a TCP packet with SYN, ACK, and Data fields.
type Packet struct {
	Syn, Ack int
	Data     string
}

func main() {
	// Generate a random sequence number for the SYN packet.
	seqNum := rand.Intn(10000)

	// Establish a TCP connection to the server.
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

	// Create and send the SYN packet with a retry mechanism.
	packet := Packet{Syn: seqNum}
	syn, err := toBytes(packet)
	if err != nil {
		fmt.Println("Error converting packet to bytes:", err)
		return
	}

	for i := 0; i < 3; i++ {
		fmt.Println("Client sending:", string(syn))
		_, err = conn.Write(syn)
		if err != nil {
			fmt.Println("Error writing to connection:", err)
			return
		}

		// Receive SYN-ACK with a timeout.
		buf := make([]byte, 1024)
		err := conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		if err != nil {
			fmt.Println("Error setting read deadline:", err)
			return
		}
		n, err := conn.Read(buf)
		if err == nil {
			synAck := toPacket(buf[:n])
			if synAck.Ack == seqNum+1 {
				fmt.Println("Client received:", string(buf[:n]))

				// Prepare and send the ACK packet.
				message := "Hello World"
				if len(os.Args) > 1 {
					message = os.Args[1]
				}
				packet = Packet{Syn: synAck.Ack, Ack: synAck.Syn + 1, Data: message}
				ack, err := toBytes(packet)
				if err != nil {
					fmt.Println("Error converting packet to bytes:", err)
					return
				}
				fmt.Println("Client sending:", string(ack))
				_, err = conn.Write(ack)
				if err != nil {
					fmt.Println("Error writing to connection:", err)
					return
				}
				return
			}
		}
		fmt.Println("Error reading from connection, retrying:", err)
	}
	fmt.Println("Failed to receive SYN-ACK after retries")
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
