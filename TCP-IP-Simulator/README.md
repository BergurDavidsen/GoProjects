#### Group: Team Burger
##### André arbi@itu.dk
##### Bergur berd@itu.dk
##### Bror broh@itu.dk
##### Konrad koed@itu.dk

*a) What are packages in your implementation? What data structure do you use to transmit data and meta-data?*

- Our solution implements a type called ***Packet***. This can include *SYN ACK* and *Data*. This is to ensure a single packet exchange between a client and server.

*b) Does your implementation use threads or processes? Why is it not realistic to use threads?*

- The implementation uses goroutines, which are lightweight, managed by the Go runtime, and used for concurrency. Each of the go statements in your code, like go handleConnection(conn), spawns a new goroutine.
The implementation doesn’t use separate OS processes. All the code runs in the same process, managed by the Go runtime, utilizing goroutines for concurrent tasks.
OS threads are much heavier than goroutines, consuming more memory and CPU resources. If you were to handle each connection with a separate thread, it would not scale well with thousands of concurrent connections.

*c) In case the network changes the order in which messages are delivered, how would you handle message re-ordering?*

- Each message is assigned a unique sequence number, and the reciever checks the sequence numbers and sorts in order before processing.

*d) In case messages can be delayed or lost, how does your implementation handle message loss?*

  - We have implemented an acknowledgement and retransmission system, where each message sent between the parties would require an acknowledgement from the receiver, within a certain time period. This would ensure a party could communicate a failure of message-transmission and order a retransmition of the message between 1-3 times; After 3 retries, it stops and prints a fail statement. 
  
*e) Why is the 3-way handshake important?*

- The main reason is to ensure a stable connection between the client and server.
  - This includes that the corresponding packets that are sent between each party are the correct ones in which they receive.
