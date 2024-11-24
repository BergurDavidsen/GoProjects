# BSDISYS1KU, Distributed Systems, BSc (Autumn 2024)
## Mandatory Activity 5 - Replication
> **Authors:**  
> André Birk \<arbi@itu.dk\>  
> Bergur Davidsen \<berd@itu.dk\>  
> Bror Hansen \<broh@itu.dk\>  
> Konrad Adolph \<koad@itu.dk\>  


### Guide
___
#### 1. Setup server
1.1 Ensure you are in the Replication root folder.  
1.2 Type `$ go run server/main.go <port>`.  
\- *If the server doesn't start, please try on another port.*  
1.3 Repeat this for as many server nodes as needed.

#### 2. Client server
2.1 Ensure you are in the Replication root folder.  
2.2 Run this command where you specify a client name and specify each server node's port.  
\- Type `$ go run client/main.go <name> <port1> <port2> <port3> ...`  
\- *if you have setup more than three server nodes up, you can feel free to add them to the end in the same format.*
