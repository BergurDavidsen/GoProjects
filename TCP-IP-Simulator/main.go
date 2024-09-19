package main

import (
	"fmt"
	"time"
)

func main() {
	clientSender := make(chan string, 1)
	clientReceiver := make(chan string, 1)
	clientVerification := make(chan string, 1)

	serverSender := make(chan string, 1)
	serverReceiver := make(chan string, 1)
	logger := make(chan string, 1)

	go client(clientSender, clientReceiver, clientVerification)
	go server(serverSender, serverReceiver)
	go forwarder(clientSender, clientReceiver, serverSender, serverReceiver)
	go God(logger)

	time.Sleep(5 * time.Second)
}

func client(sender, receiver, verification chan string) {
	seq := 1

	// SYN msg to server
	msg := getMessage(seq, "SYN")
	fmt.Println("Client sending msg:", msg)
	sender <- msg

	// Receive SYN-ACK from server
	synAck := <-receiver
	fmt.Println("Client received:", synAck)

	// ACK msg to sender channel
	msg = getMessage(seq+1, "ACK")
	fmt.Println("Client sending msg:", msg)
	sender <- msg
}

func server(sender, receiver chan string) {
	syn := <-receiver
	fmt.Println("Server received:", syn)

	msg := getMessage(1, "SYN-ACK")
	fmt.Println("Server sending msg:", msg)
	sender <- msg

	ack := <-receiver
	fmt.Println("Server received:", ack)
}

func forwarder(clientSender, clientReceiver, serverSender, serverReceiver chan string) {
	for {
		select {
		case msg := <-clientSender:
			fmt.Println("Forwarder received from client:", msg)
			serverReceiver <- msg
		case msg := <-serverSender:
			fmt.Println("Forwarder received from server:", msg)
			clientReceiver <- msg
		}
	}
}

func God(logger chan string) {
	for log := range logger {
		fmt.Println(log)
	}
}

func getMessage(seqNo int, msgType string) string {
	return fmt.Sprintf("SEQ: %d, TYPE: %s", seqNo, msgType)
}
