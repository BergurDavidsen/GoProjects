# BSDISYS1KU, Distributed Systems, BSc (Autumn 2024)
## Mandatory Activity 4 - Consensus
### Authored by Team Burger:  
> André Birk <arbi@itu.dk>  
> Bergur Davidsen <Berd@itu.dk>  
> Bror Hansen <Broh@itu.dk>  
> Konrad Meno Adolph <Koad@itu.dk>  



## How-to-use guide:   
#### 1. Run a node  
1.1 Open a terminal and navigate to the root-folder.  
1.2 Execute: 
``` bash 
$ go run node.go <nodeId> <port> <peerAddresses>
```  
This will by put the node into the network and listen for messages from other peer addresses

### 2. Example
2.1 This is an example of how you can run 3 nodes. Run each of these commands in their own terminal:
```bash
$ go run node.go 1 5001 localhost:5002 localhost:5003
$ go run node.go 2 5002 localhost:5001 localhost:5003
$ go run node.go 3 5003 localhost:5001 localhost:5002
```
