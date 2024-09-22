#### Group: Team Burger
##### André arbi@itu.dk
##### Bergur berd@itu.dk
##### Bror broh@itu.dk
##### Konrad koed@itu.dk

*a) What are packages in your implementation? What data structure do you use to transmit data and meta-data?*

- Our solution implements a type called ***Packet***. This can include *SYN ACK* and *Data*. This is to ensure a single packet exchange between a client and server.

- The data structure used to transmit data and meta-data is channels. Channels are used to transmit data and metadata between different parts of the program such as the client, server, forwarder while also providing a way for goroutines to communicate with each other. 

*b) Does your implementation use threads or processes? Why is it not realistic to use threads?*

- The implementation uses goroutines, which are lightweight threads. It's not realistic to use threads, as goroutines are more efficient in terms of memory and scheduling, specifically handling many concurrent connections. 

*c) In case the network changes the order in which messages are delivered, how would you handle message re-ordering?*

- Each message is assigned a unique sequence number, and the reciever checks the sequence numbers and sorts in order before processing.

*d) In case messages can be delayed or lost, how does your implementation handle message loss?*

- Ours currently don't have any, but if we did, this is how: 
  - We would implement an acknowledgement and retransmission system, where each message sent between the parties would require an acknowledgement from the receiver, within a certain time period. This would ensure a party could communicate a failure of message-transmission and order a retransmition of the message between 1-n times (where n is a limit); After n retries, it stops and prints a fail statement. 
  
*e) Why is the 3-way handshake important?*

- The main reason is to ensure a stable connection between the client and server.
  - This includes that the corresponding packets that are sent between each party are the correct ones in which they receive.
