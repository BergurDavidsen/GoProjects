package main

import "fmt"



func main() {
	client_sender := make(chan string, 1)
	client_receiver := make(chan string, 1)
	client_verification := make(chan string, 1)

	server_sender := make(chan string, 1)
	server_receiver := make(chan string, 1)
	server_verification := make(chan string, 1)

	God := make(chan string, 1)

	go client(client_sender, client_receiver, client_verification)
	go server(server_sender, server_receiver, server_verification)
	go forwarder(client_sender, client_receiver, client_verification, server_verification)
	go God(logger)
}

func client(sequence_number, sender, receiver, verification chan string) {
	seq := 1 //starts with 1

	//sends syn msg to sender channel
	msg := getMessage(seq, "SYN")
	fmt.Print("Client sending msg: ", msg)
	sender <- msg

	//waits to receive SYN-ACK from client
	synaAck := <-receiver
	fmt.Println("Client received: ", synAck)

	msg = getMessage(seq+1, "ACK")
	fmt.
}

func server(sequence_number, sender, receiver, verification chan string) {

}


func forwarder(client_sender, client_receiver, client_verification,
	server_sender, server_receiver, server_verification chan string) {

	for{
		select{
		case msg := <-client_sender:
			fmt.Println()
			server_receiver <- msg
		}
	}
}

func God(logger chan string){
	for log := range logger{
		fmt.Println(log)
	}
}

func verification(sender, receiver, verification chan string) {

}

func getMessage(SEQno int, Type string) string {

	return "fmt."
}

func randomSequenceNo() int {

}
